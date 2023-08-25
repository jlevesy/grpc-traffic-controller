package gtc

import (
	"context"

	resourcesv3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	gtcv1alpha1 "github.com/jlevesy/grpc-traffic-controller/api/gtc/v1alpha1"
	gtclisters "github.com/jlevesy/grpc-traffic-controller/client/listers/gtc/v1alpha1"
	"go.uber.org/zap"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type grpcListenerChangedHandler struct {
	watches *watches
	logger  *zap.Logger
}

func (h *grpcListenerChangedHandler) OnAdd(ctx context.Context, obj any) error {
	return h.handle(ctx, obj)
}

func (h *grpcListenerChangedHandler) OnUpdate(ctx context.Context, oldObj, newObj any) error {
	// TODO(jly) be clever and detect changes if it makes sense.
	return h.handle(ctx, newObj)
}

func (h *grpcListenerChangedHandler) OnDelete(ctx context.Context, obj any) error {
	return h.handle(ctx, obj)
}

func (h *grpcListenerChangedHandler) handle(ctx context.Context, obj any) error {
	lis, ok := obj.(*gtcv1alpha1.GRPCListener)
	if !ok {
		h.logger.Error("Invalid object type, expected an GRPCListener")
		return nil
	}

	h.logger.Debug(
		"gRPC Listener Changed",
		zap.String("grpc_listener_namespace", lis.GetNamespace()),
		zap.String("grcp_listener_name", lis.GetName()),
	)

	h.watches.notifyChanged(
		ctx,
		resourceRef{
			typeURL:      resourcesv3.ListenerType,
			resourceName: listenerName(lis.GetNamespace(), lis.GetName()),
		},
	)

	for routeID, route := range lis.Spec.Routes {
		for backendID := range route.Backends {
			h.watches.notifyChanged(
				ctx,
				resourceRef{
					typeURL: resourcesv3.ClusterType,
					resourceName: backendName(
						lis.GetNamespace(),
						lis.GetName(),
						routeID,
						backendID,
					),
				},
			)

			h.watches.notifyChanged(
				ctx,
				resourceRef{
					typeURL: resourcesv3.EndpointType,
					resourceName: backendName(
						lis.GetNamespace(),
						lis.GetName(),
						routeID,
						backendID,
					),
				},
			)
		}

	}

	return nil
}

type endpointSliceChangedHandler struct {
	watches *watches
	logger  *zap.Logger

	listenersLister gtclisters.GRPCListenerLister
}

func (h *endpointSliceChangedHandler) OnAdd(ctx context.Context, obj any) error {
	return h.handle(ctx, obj)
}

func (h *endpointSliceChangedHandler) OnUpdate(ctx context.Context, oldObj, newObj any) error {
	// TODO(jly) be clever and detect changes if it makes sense.
	return h.handle(ctx, newObj)
}

func (h *endpointSliceChangedHandler) OnDelete(ctx context.Context, obj any) error {
	return h.handle(ctx, obj)
}

func (h *endpointSliceChangedHandler) handle(ctx context.Context, obj any) error {
	objMeta, err := apimeta.Accessor(obj)
	if err != nil {
		h.logger.Error("Could not convert object meta", zap.Error(err))
		return err
	}

	listeners, err := h.listenersLister.List(labels.Everything())
	if err != nil {
		h.logger.Error("Could not list gRPC listeners", zap.Error(err))
		return err
	}

	// O(n) accross all services isn't good. Yet that's the price of maintaining cross namespace localities.
	// Dropping this feature would allow us to narrow down the list of services to lookup by namespace.
	for _, lis := range listeners {
		for routeID, route := range lis.Spec.Routes {
			for backendID, backend := range route.Backends {
				if matchesBackend(objMeta, lis, backend) {
					h.logger.Debug(
						"Endpoint changed",
						zap.String("grpc_listener_namespace", lis.GetNamespace()),
						zap.String("grpc_listener_name", lis.GetName()),
						zap.String("endpoint_name", objMeta.GetName()),
						zap.String("endpoint_namespace", objMeta.GetNamespace()),
					)

					h.watches.notifyChanged(
						ctx,
						resourceRef{
							typeURL: resourcesv3.EndpointType,
							resourceName: backendName(
								lis.GetNamespace(),
								lis.GetName(),
								routeID,
								backendID,
							),
						},
					)
				}
			}
		}
	}

	return nil
}

func matchesBackend(epSlice metav1.Object, listener *gtcv1alpha1.GRPCListener, backend gtcv1alpha1.Backend) bool {
	switch {
	case backend.Service != nil:
		return matchesService(epSlice, listener, backend.Service)
	case len(backend.Localities) > 0:
		for _, loc := range backend.Localities {
			if matchesService(epSlice, listener, loc.Service) {
				return true
			}
		}

		return false
	default:
		return false
	}
}

func matchesService(epSlice metav1.Object, listener *gtcv1alpha1.GRPCListener, serviceRef *gtcv1alpha1.ServiceRef) bool {
	// If the name doesn't match then we're out.
	if svcName := epSlice.GetLabels()["kubernetes.io/service-name"]; svcName != serviceRef.Name {
		return false
	}

	// If we don't have a specific namespace, then we match looking at the XDS service namespace.
	return listener.Namespace == epSlice.GetNamespace() || (serviceRef.Namespace != "" && serviceRef.Namespace == epSlice.GetNamespace())
}
