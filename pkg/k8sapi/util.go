package k8sapi

import (
	"context"
	"fmt"
	"strings"
	"sync"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	argoRollouts "github.com/datawire/argo-rollouts-go-client/pkg/client/clientset/versioned"
	"github.com/datawire/dlib/dlog"
)

func WithJoinedClientSetInterface(ctx context.Context, ki kubernetes.Interface, ari argoRollouts.Interface) context.Context {
	return WithArgoRolloutsInterface(WithK8sInterface(ctx, ki), ari)
}

func GetJoinedClientSetInterface(ctx context.Context) JoinedClientSetInterface {
	return &joinedClientSetInterface{
		GetK8sInterface(ctx),
		GetArgoRolloutsInterface(ctx),
	}
}

func WithArgoRolloutsInterface(ctx context.Context, ari argoRollouts.Interface) context.Context {
	return context.WithValue(ctx, ariKey{}, ari)
}

func WithK8sInterface(ctx context.Context, ki kubernetes.Interface) context.Context {
	return context.WithValue(ctx, kiKey{}, ki)
}

func GetArgoRolloutsInterface(ctx context.Context) argoRollouts.Interface {
	ari, ok := ctx.Value(ariKey{}).(argoRollouts.Interface)
	if !ok {
		return nil
	}
	return ari
}

func GetK8sInterface(ctx context.Context) kubernetes.Interface {
	ki, ok := ctx.Value(kiKey{}).(kubernetes.Interface)
	if !ok {
		panic("K8sInterface requested from a context that has none")
	}
	return ki
}

type kiKey struct{}

type ariKey struct{}

// GetPort finds a port with the given name and returns it.
func GetPort(cn *core.Container, portName string) (*core.ContainerPort, error) {
	ports := cn.Ports
	for pn := range ports {
		p := &ports[pn]
		if p.Name == portName {
			return p, nil
		}
	}
	return nil, fmt.Errorf("unable to locate port %q in container %q", portName, cn.Name)
}

// GetAppProto determines the application protocol of the given ServicePort. The given AppProtocolStrategy
// used if the port's appProtocol attribute is unset.
func GetAppProto(ctx context.Context, aps AppProtocolStrategy, p *core.ServicePort) string {
	if p.AppProtocol != nil {
		appProto := *p.AppProtocol
		if appProto != "" {
			dlog.Debugf(ctx, "Using application protocol %q from service appProtocol field", appProto)
			return appProto
		}
	}

	switch aps {
	case Http:
		return "http"
	case Http2:
		return "http2"
	case PortName:
		if p.Name == "" {
			dlog.Debug(ctx, "Unable to derive application protocol from unnamed service port with no appProtocol field")
			break
		}
		pn := p.Name
		if dashPos := strings.IndexByte(pn, '-'); dashPos > 0 {
			pn = pn[:dashPos]
		}
		var appProto string
		switch strings.ToLower(pn) {
		case "http", "https", "grpc", "http2":
			appProto = pn
		case "h2c": // h2c is cleartext HTTP/2
			appProto = "http2"
		case "tls", "h2": // same as https in this context and h2 is HTTP/2 with TLS
			appProto = "https"
		}
		if appProto != "" {
			dlog.Debugf(ctx, "Using application protocol %q derived from port name %q", appProto, p.Name)
			return appProto
		}
		dlog.Debugf(ctx, "Unable to derive application protocol from port name %q", p.Name)
	}
	return ""
}

func ObjErrorf(o Object, format string, args ...any) error {
	return fmt.Errorf("%s name=%q namespace=%q: %w",
		o.GetKind(), o.GetName(), o.GetNamespace(),
		fmt.Errorf(format, args...))
}

func listOptions(labelSelector labels.Set) meta.ListOptions {
	opts := meta.ListOptions{}
	if len(labelSelector) > 0 {
		opts.LabelSelector = labels.SelectorFromSet(labelSelector).String()
	}
	return opts
}

// Subscribe writes to the given channel whenever relevant information has changed
// in the current snapshot.
func Subscribe(c context.Context, cond *sync.Cond) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		for {
			cond.L.Lock()
			cond.Wait()
			cond.L.Unlock()

			select {
			case <-c.Done():
				close(ch)
				return
			case ch <- struct{}{}:
			}
		}
	}()
	return ch
}
