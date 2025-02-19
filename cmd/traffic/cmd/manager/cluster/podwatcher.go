package cluster

import (
	"context"
	"math"
	"net/netip"
	"slices"
	"sync"
	"time"

	"github.com/puzpuzpuz/xsync/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	"github.com/datawire/dlib/dlog"
	"github.com/telepresenceio/telepresence/v2/cmd/traffic/cmd/manager/namespaces"
	"github.com/telepresenceio/telepresence/v2/pkg/informer"
	"github.com/telepresenceio/telepresence/v2/pkg/subnet"
)

// PodLister helps list Pods.
// All objects returned here must be treated as read-only.
type PodLister interface {
	// List lists all Pods in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*corev1.Pod, err error)
}

type podWatcher struct {
	ipsMap    map[netip.Addr]struct{}
	timer     *time.Timer
	notifyCh  chan subnet.Set
	informers *xsync.MapOf[string, cache.ResourceEventHandlerRegistration]
	lock      sync.Mutex // Protects all access to ipsMap
}

func newPodWatcher(ctx context.Context, managerIP netip.Addr) *podWatcher {
	w := &podWatcher{
		ipsMap:    make(map[netip.Addr]struct{}),
		notifyCh:  make(chan subnet.Set),
		informers: xsync.NewMapOf[string, cache.ResourceEventHandlerRegistration](),
	}
	w.ipsMap[managerIP] = struct{}{}

	var oldSubnets subnet.Set
	sendIfChanged := func() {
		w.lock.Lock()
		ips := make([]netip.Addr, len(w.ipsMap))
		i := 0
		for ip := range w.ipsMap {
			ips[i] = ip
			i++
		}
		w.lock.Unlock()

		newSubnets := subnet.NewSet(subnet.CoveringPrefixes(ips))
		if !newSubnets.Equals(oldSubnets) {
			dlog.Debugf(ctx, "podWatcher calling updateSubnets with %v", newSubnets)
			select {
			case <-ctx.Done():
				return
			case w.notifyCh <- newSubnets:
				oldSubnets = newSubnets
			}
		}
	}

	w.timer = time.AfterFunc(time.Duration(math.MaxInt64), sendIfChanged)
	go func() {
		id, nsChanges := namespaces.Subscribe(ctx)
		defer namespaces.Unsubscribe(ctx, id)
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-nsChanges:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					return
				default:
					w.refreshWatchers(ctx)
				}
			}
		}
	}()
	return w
}

func (w *podWatcher) refreshWatchers(ctx context.Context) {
	nss := namespaces.GetOrGlobal(ctx)

	// Register event handlers for namespaces that are no longer managed
	for _, ns := range nss {
		w.informers.Compute(ns, func(reg cache.ResourceEventHandlerRegistration, loaded bool) (cache.ResourceEventHandlerRegistration, bool) {
			if loaded {
				return reg, false
			}
			inf := informer.GetK8sFactory(ctx, ns).Core().V1().Pods().Informer()
			reg, err := inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj any) {
					if pod, ok := obj.(*corev1.Pod); ok {
						w.onPodAdded(ctx, pod)
					}
				},
				DeleteFunc: func(obj any) {
					if pod, ok := obj.(*corev1.Pod); ok {
						w.onPodDeleted(ctx, pod)
					} else if dfsu, ok := obj.(*cache.DeletedFinalStateUnknown); ok {
						if pod, ok := dfsu.Obj.(*corev1.Pod); ok {
							w.onPodDeleted(ctx, pod)
						}
					}
				},
				UpdateFunc: func(oldObj, newObj any) {
					if oldPod, ok := oldObj.(*corev1.Pod); ok {
						if newPod, ok := newObj.(*corev1.Pod); ok {
							w.onPodUpdated(ctx, oldPod, newPod)
						}
					}
				},
			})
			if err != nil {
				dlog.Errorf(ctx, "failed to add pod watcher %q : %v", ns, err)
				return nil, true
			}
			dlog.Debugf(ctx, "add pod watcher %q", ns)
			return reg, false
		})
	}

	// Unregister event handlers for namespaces that are no longer managed
	w.informers.Range(func(ns string, reg cache.ResourceEventHandlerRegistration) bool {
		if !slices.Contains(nss, ns) {
			err := informer.GetK8sFactory(ctx, ns).Core().V1().Pods().Informer().RemoveEventHandler(reg)
			if err != nil {
				dlog.Errorf(ctx, "failed to remove pod watcher %q : %v", ns, err)
			} else {
				dlog.Debugf(ctx, "removed pod watcher %q", ns)
			}
		}
		return true
	})
}

func (w *podWatcher) changeNotifier(ctx context.Context, updateSubnets func(set subnet.Set)) {
	for {
		select {
		case <-ctx.Done():
			return
		case subnets := <-w.notifyCh:
			updateSubnets(subnets)
		}
	}
}

func (w *podWatcher) viable(ctx context.Context) bool {
	w.lock.Lock()
	defer w.lock.Unlock()
	if len(w.ipsMap) > 0 {
		return true
	}

	nss := namespaces.GetOrGlobal(ctx)

	// Create the initial snapshot
	var pods []*corev1.Pod
	var err error
	for _, ns := range nss {
		lister := informer.GetK8sFactory(ctx, ns).Core().V1().Pods().Lister()
		if ns != "" {
			pods, err = lister.Pods(ns).List(labels.Everything())
		} else {
			pods, err = lister.List(labels.Everything())
		}
		if err != nil {
			dlog.Errorf(ctx, "unable to list pods: %v", err)
			return false
		}
		for _, pod := range pods {
			w.addLocked(podIPs(ctx, pod))
		}
	}

	return true
}

func (w *podWatcher) onPodAdded(ctx context.Context, pod *corev1.Pod) {
	if ipKeys := podIPs(ctx, pod); len(ipKeys) > 0 {
		w.add(ipKeys)
	}
}

func (w *podWatcher) onPodDeleted(ctx context.Context, pod *corev1.Pod) {
	if ipKeys := podIPs(ctx, pod); len(ipKeys) > 0 {
		w.drop(ipKeys)
	}
}

func (w *podWatcher) onPodUpdated(ctx context.Context, oldPod, newPod *corev1.Pod) {
	added, dropped := getIPsDelta(podIPs(ctx, oldPod), podIPs(ctx, newPod))
	if len(added) > 0 {
		if len(dropped) > 0 {
			w.update(dropped, added)
		} else {
			w.add(added)
		}
	} else if len(dropped) > 0 {
		w.drop(dropped)
	}
}

const podWatcherSendDelay = 10 * time.Millisecond

func (w *podWatcher) add(ips []netip.Addr) {
	w.lock.Lock()
	w.addLocked(ips)
	w.lock.Unlock()
}

func (w *podWatcher) drop(ips []netip.Addr) {
	w.lock.Lock()
	w.dropLocked(ips)
	w.lock.Unlock()
}

func (w *podWatcher) update(dropped, added []netip.Addr) {
	w.lock.Lock()
	w.dropLocked(dropped)
	w.addLocked(added)
	w.lock.Unlock()
}

func (w *podWatcher) addLocked(ips []netip.Addr) {
	if w.ipsMap == nil {
		w.ipsMap = make(map[netip.Addr]struct{}, 100)
	}

	changed := false
	exists := struct{}{}
	for _, ip := range ips {
		if _, ok := w.ipsMap[ip]; !ok {
			w.ipsMap[ip] = exists
			changed = true
		}
	}
	if changed {
		w.timer.Reset(podWatcherSendDelay)
	}
}

func (w *podWatcher) dropLocked(ips []netip.Addr) {
	changed := false
	for _, ip := range ips {
		if _, ok := w.ipsMap[ip]; ok {
			delete(w.ipsMap, ip)
			changed = true
		}
	}
	if changed {
		w.timer.Reset(podWatcherSendDelay)
	}
}

// getIPsDelta returns the difference between the old and new IPs.
//
// NOTE! The array of the old slice is modified and used for the dropped return.
func getIPsDelta(oldIPs, newIPs []netip.Addr) (added, dropped []netip.Addr) {
	lastOI := len(oldIPs) - 1
	if lastOI < 0 {
		return newIPs, nil
	}

nextN:
	for _, n := range newIPs {
		for oi, o := range oldIPs {
			if n == o {
				oldIPs[oi] = oldIPs[lastOI]
				oldIPs = oldIPs[:lastOI]
				lastOI--
				continue nextN
			}
		}
		added = append(added, n)
	}
	if len(oldIPs) == 0 {
		oldIPs = nil
	}
	return added, oldIPs
}

func podIPs(ctx context.Context, pod *corev1.Pod) []netip.Addr {
	if pod == nil {
		return nil
	}
	if pod.Namespace == "kube-system" {
		// If the user wants the pod subnet of this namespace mapped, they'll need to add it manually.
		// Auto-generating it here will often cause problems. Especially when running Kubernetes locally.
		return nil
	}
	status := pod.Status
	podIPs := status.PodIPs
	if len(podIPs) == 0 {
		if status.PodIP == "" {
			return nil
		}
		podIPs = []corev1.PodIP{{IP: status.PodIP}}
	}
	ips := make([]netip.Addr, 0, len(podIPs))
	for _, ps := range podIPs {
		ip, err := netip.ParseAddr(ps.IP)
		if err != nil {
			dlog.Errorf(ctx, "unable to parse IP %q in pod %s.%s", ps.IP, pod.Name, pod.Namespace)
			continue
		}
		ips = append(ips, ip)
	}
	return ips
}
