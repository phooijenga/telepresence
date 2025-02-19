package mutator

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	admission "k8s.io/api/admission/v1"
	core "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/strings/slices"

	"github.com/datawire/dlib/derror"
	"github.com/datawire/dlib/dlog"
	"github.com/telepresenceio/telepresence/v2/cmd/traffic/cmd/manager/managerutil"
	"github.com/telepresenceio/telepresence/v2/pkg/agentconfig"
	"github.com/telepresenceio/telepresence/v2/pkg/agentmap"
	"github.com/telepresenceio/telepresence/v2/pkg/k8sapi"
	"github.com/telepresenceio/telepresence/v2/pkg/maps"
	"github.com/telepresenceio/telepresence/v2/pkg/workload"
)

var podResource = meta.GroupVersionResource{Version: "v1", Group: "", Resource: "pods"} //nolint:gochecknoglobals // constant

type AgentInjector interface {
	Inject(ctx context.Context, req *admission.AdmissionRequest) (p PatchOps, err error)
	Uninstall(ctx context.Context)
}

// NewAgentInjector creates a new agentInjector.
func NewAgentInjector(_ context.Context, agentConfigs Map) AgentInjector {
	ai := &agentInjector{
		agentConfigs: agentConfigs,
	}
	return ai
}

var NewAgentInjectorFunc = NewAgentInjector //nolint:gochecknoglobals // extension point

type agentInjector struct {
	sync.Mutex
	agentConfigs Map
	terminating  int64
}

func getPod(req *admission.AdmissionRequest, isDelete bool) (*core.Pod, error) {
	if req.Resource != podResource {
		return nil, fmt.Errorf("expect resource to be %s, got %s", podResource, req.Resource)
	}

	// Parse the Pod object.
	var raw []byte
	if isDelete {
		raw = req.OldObject.Raw
	} else {
		raw = req.Object.Raw
	}
	pod := core.Pod{}
	if _, _, err := universalDeserializer.Decode(raw, nil, &pod); err != nil {
		return nil, fmt.Errorf("could not deserialize pod object: %v", err)
	}

	podNamespace := pod.Namespace
	if podNamespace == "" {
		// It is very probable the pod was not yet assigned a namespace,
		// in which case we should use the AdmissionRequest namespace.
		pod.Namespace = req.Namespace
	}
	podName := pod.Name
	if podName == "" {
		// It is very probable the pod was not yet assigned a name,
		// in which case we should use the metadata generated name.
		pod.Name = pod.ObjectMeta.GenerateName
	}

	// Validate traffic-agent injection preconditions.
	if pod.Name == "" || pod.Namespace == "" {
		return nil, fmt.Errorf(`unable to extract pod name and/or namespace (got "%s.%s")`, pod.Name, pod.Namespace)
	}
	return &pod, nil
}

func (a *agentInjector) Inject(ctx context.Context, req *admission.AdmissionRequest) (p PatchOps, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = derror.PanicToError(r)
			dlog.Errorf(ctx, "%+v", err)
		}
	}()

	isDelete := req.Operation == admission.Delete
	if atomic.LoadInt64(&a.terminating) > 0 {
		dlog.Debugf(ctx, "Skipping webhook for %s.%s because the agent-injector is terminating", req.Name, req.Namespace)
		return nil, nil
	}

	pod, err := getPod(req, isDelete)
	if err != nil {
		return nil, err
	}

	if isDelete {
		a.agentConfigs.Blacklist(pod.Name, pod.Namespace)
		return nil, nil
	}

	dlog.Debugf(ctx, "Handling admission request %s %s.%s", req.Operation, pod.Name, pod.Namespace)
	env := managerutil.GetEnv(ctx)

	ia := pod.Annotations[agentconfig.InjectAnnotation]

	var scx agentconfig.SidecarExt
	switch ia {
	case "false", "disabled":
		dlog.Debugf(ctx, `The %s.%s pod is explicitly disabled using a %q annotation; skipping`, pod.Name, pod.Namespace, agentconfig.InjectAnnotation)
		return nil, nil
	case "":
		if env.AgentInjectPolicy != agentconfig.OnDemand {
			dlog.Debugf(ctx, `The %s.%s pod has not enabled %s container injection through %q annotation; skipping`,
				pod.Name, pod.Namespace, agentconfig.ContainerName, agentconfig.InjectAnnotation)
			return nil, nil
		}
		fallthrough
	case "enabled":
		img := managerutil.GetAgentImage(ctx)
		if img == "" {
			dlog.Debug(ctx, "Skipping webhook injection because the traffic-manager is unable to determine what image to use for injected traffic-agents.")
			return nil, nil
		}

		enabledWorkloads := managerutil.GetEnv(ctx).EnabledWorkloadKinds
		supportedKinds := make([]string, len(enabledWorkloads))
		for i, wlKind := range enabledWorkloads {
			switch wlKind {
			case workload.DeploymentKind:
				supportedKinds[i] = "Deployment"
			case workload.ReplicaSetKind:
				supportedKinds[i] = "ReplicaSet"
			case workload.StatefulSetKind:
				supportedKinds[i] = "StatefulSet"
			case workload.RolloutKind:
				supportedKinds[i] = "Rollout"
			}
		}
		wl, err := agentmap.FindOwnerWorkload(ctx, k8sapi.Pod(pod), supportedKinds)
		if err != nil {
			uwkError := k8sapi.UnsupportedWorkloadKindError("")
			switch {
			case k8sErrors.IsNotFound(err):
				dlog.Debugf(ctx, "No workload owner found for pod %s.%s", pod.Name, pod.Namespace)
			case errors.As(err, &uwkError):
				dlog.Debugf(ctx, "Workload owner with %s found for pod %s.%s", uwkError.Error(), pod.Name, pod.Namespace)
			default:
				dlog.Debugf(ctx, "No workload owner found for pod %s.%s: %v", pod.Name, pod.Namespace, err)
			}
			// Not an error. It just means that the pod is not eligible for intercepts.
			return nil, nil
		}
		scx, err = a.agentConfigs.Get(ctx, wl.GetName(), wl.GetNamespace())
		switch {
		case err != nil:
			dlog.Errorf(ctx, "Failed to retrieve agent config for workload %s.%s: %v", wl.GetName(), wl.GetNamespace(), err)
			return nil, err
		case scx == nil:
			if ia == "enabled" {
				// A race condition may occur when a workload with "enabled" is applied.
				// The workload event handler will create the agent config, but the webhook injection call may arrive before
				// that agent config has been created. When this happens, we must force a second event when the creation
				// is complete.
				a.agentConfigs.registerPrematureInjectEvent(wl)
				dlog.Debugf(ctx, "No agent config has been generated for annotation enabled %s.%s", wl.GetName(), wl.GetNamespace())
			} else {
				dlog.Debugf(ctx, "Skipping %s.%s (no agent config)", wl.GetName(), wl.GetNamespace())
			}
			return nil, nil
		case scx.AgentConfig().Manual:
			dlog.Debugf(ctx, "Skipping webhook where agent is manually injected %s.%s", wl.GetName(), wl.GetNamespace())
			return nil, nil
		}
	default:
		return nil, fmt.Errorf("invalid value %q for annotation %s", ia, agentconfig.InjectAnnotation)
	}

	var patches PatchOps
	var annotations map[string]string
	config := scx.AgentConfig()
	patches = addInitContainer(pod, config, patches)
	patches, annotations = addAgentContainer(ctx, pod, config, patches)
	patches = addPullSecrets(pod, config, patches)
	patches = addAgentVolumes(pod, config, patches)
	patches = hidePorts(pod, config, patches)
	annotations[agentconfig.InjectAnnotation] = "enabled"
	patches = addPodAnnotations(pod, annotations, patches)
	patches = addPodLabels(ctx, pod, config, patches)
	patches = maybeRemoveAppContainer(pod, config, patches)

	if config.APIPort != 0 {
		tpEnv := make(map[string]string)
		tpEnv[agentconfig.EnvAPIPort] = strconv.Itoa(int(config.APIPort))
		patches = addTPEnv(pod, config, tpEnv, patches)
	}

	// Create patch operations to add the traffic-agent sidecar
	if len(patches) > 0 {
		dlog.Infof(ctx, "Injecting %d patches into pod %s.%s", len(patches), pod.Name, pod.Namespace)
		if dlog.MaxLogLevel(ctx) >= dlog.LogLevelDebug {
			cns := strings.Builder{}
			for i, cn := range pod.Spec.Containers {
				cns.WriteString(fmt.Sprintf("%d %s\n", i, cn.Name))
			}
			dlog.Debugf(ctx, "Containers \n%s", cns.String())
			if pj, err := json.Marshal(patches, jsontext.WithIndent("  ")); err == nil {
				dlog.Debugf(ctx, "\n%s", string(pj))
			}
		}
	} else {
		dlog.Infof(ctx, "Pod %s.%s was left untouched", pod.Name, pod.Namespace)
	}
	return patches, nil
}

// Uninstall ensures that no more webhook injections are made and that all the workloads of currently injected
// pods are rolled out.
func (a *agentInjector) Uninstall(ctx context.Context) {
	atomic.StoreInt64(&a.terminating, 1)
	a.agentConfigs.DeleteMapsAndRolloutAll(ctx)
}

func needInitContainer(config *agentconfig.Sidecar) bool {
	for _, cc := range config.Containers {
		if !cc.Replace {
			for _, ic := range cc.Intercepts {
				if ic.Headless || ic.TargetPortNumeric {
					return true
				}
			}
		}
	}
	return false
}

func maybeRemoveAppContainer(pod *core.Pod, config *agentconfig.Sidecar, patches PatchOps) PatchOps {
	// Extremely important to remove in reverse order, or one may affect the index of the next removal.
	cns := pod.Spec.Containers
	for i := len(cns) - 1; i >= 0; i-- {
		for _, cc := range config.Containers {
			if cc.Name == cns[i].Name && cc.Replace {
				patches = append(patches, PatchOperation{
					Op:   "remove",
					Path: fmt.Sprintf("/spec/containers/%d", i),
				})
			}
		}
	}
	return patches
}

func addInitContainer(pod *core.Pod, config *agentconfig.Sidecar, patches PatchOps) PatchOps {
	if !needInitContainer(config) {
		for i, oc := range pod.Spec.InitContainers {
			if agentconfig.InitContainerName == oc.Name {
				return append(patches, PatchOperation{
					Op:   "remove",
					Path: fmt.Sprintf("/spec/initContainers/%d", i),
				})
			}
		}
		return patches
	}

	pis := pod.Spec.InitContainers
	ic := agentconfig.InitContainer(config)
	if len(pis) == 0 {
		return append(patches, PatchOperation{
			Op:    "replace",
			Path:  "/spec/initContainers",
			Value: []core.Container{*ic},
		})
	}

	for i := range pis {
		oc := &pis[i]
		if ic.Name == oc.Name {
			if ic.Image == oc.Image &&
				slices.Equal(ic.Args, oc.Args) &&
				compareVolumeMounts(ic.VolumeMounts, oc.VolumeMounts) &&
				compareCapabilities(ic.SecurityContext, oc.SecurityContext) {
				return patches
			}
			return append(patches, PatchOperation{
				Op:    "replace",
				Path:  fmt.Sprintf("/spec/initContainers/%d", i),
				Value: ic,
			})
		}
	}

	return append(patches, PatchOperation{
		Op:    "add",
		Path:  "/spec/initContainers/-",
		Value: ic,
	})
}

func addAgentVolumes(pod *core.Pod, ag *agentconfig.Sidecar, patches PatchOps) PatchOps {
	for _, vol := range pod.Spec.Volumes {
		if vol.Name == agentconfig.AnnotationVolumeName {
			return patches
		}
	}
	avs := agentconfig.AgentVolumes(ag.AgentName, pod)
	if len(avs) == 0 {
		return patches
	}

	// Ensure that /spec/volumes exists in the pod. It won't be present when the pod doesn't have
	// any volumes and automountServiceAccountToken == false
	if pod.Spec.Volumes == nil {
		patches = append(patches,
			PatchOperation{
				Op:    "replace",
				Path:  "/spec/volumes",
				Value: avs,
			})
	} else {
		for _, av := range avs {
			patches = append(patches,
				PatchOperation{
					Op:    "add",
					Path:  "/spec/volumes/-",
					Value: av,
				})
		}
	}
	return patches
}

// compareProbes compares two Probes but will only consider their Handler.Exec.Command in the comparison.
func compareProbes(a, b *core.Probe) bool {
	if a == nil || b == nil {
		return a == b
	}
	ae := a.ProbeHandler.Exec
	be := b.ProbeHandler.Exec
	if ae == nil || be == nil {
		return ae == be
	}
	eq := cmp.Equal(ae.Command, be.Command)
	return eq
}

func compareCapabilities(a *core.SecurityContext, b *core.SecurityContext) bool {
	ac := a.Capabilities
	bc := b.Capabilities
	if ac == bc {
		return true
	}
	if ac == nil || bc == nil {
		return false
	}
	compareCaps := func(acs []core.Capability, bcs []core.Capability) bool {
		if len(acs) != len(bcs) {
			return false
		}
		for i := range acs {
			if acs[i] != bcs[i] {
				return false
			}
		}
		return true
	}
	return compareCaps(ac.Add, bc.Add) && compareCaps(ac.Drop, bc.Drop)
}

// compareVolumeMounts compares two VolumeMount slices but will not include volume mounts using "kube-api-access-" prefix.
func compareVolumeMounts(a, b []core.VolumeMount) bool {
	stripKubeAPI := func(vs []core.VolumeMount) []core.VolumeMount {
		ss := make([]core.VolumeMount, 0, len(vs))
		for _, v := range vs {
			if !(strings.HasPrefix(v.Name, "kube-api-access-") || strings.HasPrefix(v.MountPath, "/var/run/secrets/kubernetes.io/")) {
				ss = append(ss, v)
			}
		}
		return ss
	}
	eq := cmp.Equal(stripKubeAPI(a), stripKubeAPI(b))
	return eq
}

func containerEqual(ctx context.Context, a, b *core.Container) bool {
	// skips contain defaults assigned by Kubernetes that are not zero values
	options := cmp.Options{
		cmp.Comparer(compareProbes),
		cmp.Comparer(compareVolumeMounts),
		cmpopts.IgnoreFields(core.Container{}, "ImagePullPolicy", "Resources", "TerminationMessagePath", "TerminationMessagePolicy"),
	}
	if dlog.MaxLogLevel(ctx) >= dlog.LogLevelDebug {
		diff := cmp.Diff(a, b, options...)
		if diff != "" {
			dlog.Debug(ctx, diff)
		}
		return diff == ""
	}
	return cmp.Equal(a, b, options...)
}

// addAgentContainer creates a patch operation to add the traffic-agent container.
func addAgentContainer(
	ctx context.Context,
	pod *core.Pod,
	config *agentconfig.Sidecar,
	patches PatchOps,
) (PatchOps, map[string]string) {
	acn, replaceAnnotations := agentconfig.AgentContainer(ctx, pod, config)
	if acn == nil {
		return patches, replaceAnnotations
	}

	refPodName := pod.Name + "." + pod.Namespace
	for i := range pod.Spec.Containers {
		pcn := &pod.Spec.Containers[i]
		if pcn.Name == agentconfig.ContainerName {
			if containerEqual(ctx, pcn, acn) {
				dlog.Infof(ctx, "Pod %s already has container %s and it isn't modified", refPodName, agentconfig.ContainerName)
				return patches, replaceAnnotations
			}
			dlog.Debugf(ctx, "Pod %s already has container %s but it is modified", refPodName, agentconfig.ContainerName)
			return append(patches, PatchOperation{
				Op:    "replace",
				Path:  "/spec/containers/" + strconv.Itoa(i),
				Value: acn,
			}), replaceAnnotations
		}
	}

	return append(patches, PatchOperation{
		Op:    "add",
		Path:  "/spec/containers/-",
		Value: acn,
	}), replaceAnnotations
}

// addAgentContainer creates a patch operation to add the traffic-agent container.
func addPullSecrets(
	pod *core.Pod,
	config *agentconfig.Sidecar,
	patches PatchOps,
) PatchOps {
	if len(config.PullSecrets) == 0 {
		return patches
	}
	if len(pod.Spec.ImagePullSecrets) == 0 {
		return append(patches, PatchOperation{
			Op:    "replace",
			Path:  "/spec/imagePullSecrets",
			Value: config.PullSecrets,
		})
	}
	for _, nps := range config.PullSecrets {
		found := false
		for _, ips := range pod.Spec.ImagePullSecrets {
			if nps.Name == ips.Name {
				found = true
				break
			}
		}
		if !found {
			patches = append(patches, PatchOperation{
				Op:    "add",
				Path:  "/spec/imagePullSecrets/-",
				Value: nps,
			})
		}
	}
	return patches
}

// addTPEnv adds telepresence specific environment variables to all interceptable app containers.
func addTPEnv(pod *core.Pod, config *agentconfig.Sidecar, env map[string]string, patches PatchOps) PatchOps {
	agentconfig.EachContainer(pod, config, func(app *core.Container, cc *agentconfig.Container) {
		if !cc.Replace {
			patches = addContainerTPEnv(pod, app, env, patches)
		}
	})
	return patches
}

// addContainerTPEnv adds telepresence specific environment variables to the app container.
func addContainerTPEnv(pod *core.Pod, cn *core.Container, env map[string]string, patches PatchOps) PatchOps {
	if l := len(cn.Env); l > 0 {
		for _, e := range cn.Env {
			if e.ValueFrom == nil && env[e.Name] == e.Value {
				delete(env, e.Name)
			}
		}
	}
	if len(env) == 0 {
		return patches
	}
	cns := pod.Spec.Containers
	var containerPath string
	for i := range cns {
		if &cns[i] == cn {
			containerPath = fmt.Sprintf("/spec/containers/%d", i)
			break
		}
	}
	keys := make([]string, len(env))
	i := 0
	for k := range env {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	if cn.Env == nil {
		patches = append(patches, PatchOperation{
			Op:    "replace",
			Path:  fmt.Sprintf("%s/%s", containerPath, "env"),
			Value: []core.EnvVar{},
		})
	}
	for _, k := range keys {
		patches = append(patches, PatchOperation{
			Op:   "add",
			Path: fmt.Sprintf("%s/%s", containerPath, "env/-"),
			Value: core.EnvVar{
				Name:      k,
				Value:     env[k],
				ValueFrom: nil,
			},
		})
	}
	return patches
}

// hidePorts  will replace the symbolic name of a container port with a generated name. It will perform
// the same replacement on all references to that port from the probes of the container.
func hidePorts(pod *core.Pod, config *agentconfig.Sidecar, patches PatchOps) PatchOps {
	agentconfig.EachContainer(pod, config, func(app *core.Container, cc *agentconfig.Container) {
		if !cc.Replace {
			for _, ic := range agentconfig.PortUniqueIntercepts(cc) {
				if ic.Headless || ic.TargetPortNumeric {
					// Rely on iptables mapping instead of port renames
					continue
				}
				patches = hideContainerPorts(pod, app, bool(cc.Replace), ic.ContainerPortName, patches)
			}
		}
	})
	return patches
}

// hideContainerPorts  will replace the symbolic name of a container port with a generated name. It will perform
// the same replacement on all references to that port from the probes of the container.
func hideContainerPorts(pod *core.Pod, app *core.Container, isReplace bool, portName string, patches PatchOps) PatchOps {
	cns := pod.Spec.Containers
	var containerPath string
	for i := range cns {
		if &cns[i] == app {
			containerPath = fmt.Sprintf("/spec/containers/%d", i)
			break
		}
	}

	hiddenPortName := hiddenPortName(portName, 0)
	hidePort := func(path string) {
		patches = append(patches, PatchOperation{
			Op:    "replace",
			Path:  fmt.Sprintf("%s/%s", containerPath, path),
			Value: hiddenPortName,
		})
	}

	for i, p := range app.Ports {
		if p.Name == portName {
			hidePort(fmt.Sprintf("ports/%d/name", i))
			break
		}
	}

	// A replacing intercept will swap the app-container for one that doesn't have any
	// probes, so the patch must not contain renames for those.
	if !isReplace {
		probes := []*core.Probe{app.LivenessProbe, app.ReadinessProbe, app.StartupProbe}
		probeNames := []string{"livenessProbe/", "readinessProbe/", "startupProbe/"}

		for i, probe := range probes {
			if probe == nil {
				continue
			}
			if h := probe.HTTPGet; h != nil && h.Port.StrVal == portName {
				hidePort(probeNames[i] + "httpGet/port")
			}
			if t := probe.TCPSocket; t != nil && t.Port.StrVal == portName {
				hidePort(probeNames[i] + "tcpSocket/port")
			}
		}
	}
	return patches
}

func addPodAnnotations(pod *core.Pod, anns map[string]string, patches PatchOps) PatchOps {
	op := "replace"
	changed := false
	am := pod.Annotations
	if am == nil {
		op = "add"
		am = make(map[string]string)
	} else {
		am = maps.Copy(am)
	}

	for k, v := range anns {
		if _, ok := pod.Annotations[k]; !ok {
			changed = true
			am[k] = v
		}
	}

	if changed {
		patches = append(patches, PatchOperation{
			Op:    op,
			Path:  "/metadata/annotations",
			Value: am,
		})
	}
	return patches
}

func addPodLabels(_ context.Context, pod *core.Pod, config agentconfig.SidecarExt, patches PatchOps) PatchOps {
	op := "replace"
	changed := false
	lm := pod.Labels
	if lm == nil {
		op = "add"
		lm = make(map[string]string)
	} else {
		lm = maps.Copy(lm)
	}
	if _, ok := pod.Labels[agentconfig.WorkloadNameLabel]; !ok {
		changed = true
		lm[agentconfig.WorkloadNameLabel] = config.AgentConfig().WorkloadName
	}
	if _, ok := pod.Labels[agentconfig.WorkloadKindLabel]; !ok {
		changed = true
		lm[agentconfig.WorkloadKindLabel] = config.AgentConfig().WorkloadKind
	}
	if _, ok := pod.Labels[agentconfig.WorkloadEnabledLabel]; !ok {
		changed = true
		lm[agentconfig.WorkloadEnabledLabel] = "true"
	}
	if changed {
		patches = append(patches, PatchOperation{
			Op:    op,
			Path:  "/metadata/labels",
			Value: lm,
		})
	}
	return patches
}

const maxPortNameLen = 15

// hiddenPortName prefixes the given name with "tm-" and truncates it to 15 characters. If
// the ordinal is greater than zero, the last two digits are reserved for the hexadecimal
// representation of that ordinal.
func hiddenPortName(name string, ordinal int) string {
	// New name must be max 15 characters long
	hiddenName := "tm-" + name
	if len(hiddenName) > maxPortNameLen {
		if ordinal > 0 {
			hiddenName = hiddenName[:maxPortNameLen-2] + strconv.FormatInt(int64(ordinal), 16) // we don't expect more than 256 ports
		} else {
			hiddenName = hiddenName[:maxPortNameLen]
		}
	}
	return hiddenName
}
