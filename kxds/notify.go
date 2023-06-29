package kxds

import (
	"context"

	resourcesv3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	kxdsv1alpha1 "github.com/jlevesy/kxds/api/kxds/v1alpha1"
	kxdslisters "github.com/jlevesy/kxds/client/listers/kxds/v1alpha1"
	"go.uber.org/zap"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type xdsServiceChangedHandler struct {
	watches *watches
	logger  *zap.Logger
}

func (h *xdsServiceChangedHandler) OnAdd(ctx context.Context, obj any) error {
	return h.handle(ctx, obj)
}

func (h *xdsServiceChangedHandler) OnUpdate(ctx context.Context, oldObj, newObj any) error {
	// TODO(jly) be clever and detect changes if it makes sense.
	return h.handle(ctx, newObj)
}

func (h *xdsServiceChangedHandler) OnDelete(ctx context.Context, obj any) error {
	return h.handle(ctx, obj)
}

func (h *xdsServiceChangedHandler) handle(ctx context.Context, obj any) error {
	svc, ok := obj.(*kxdsv1alpha1.XDSService)
	if !ok {
		h.logger.Error("Invalid object type, expected an XDSService")
		return nil
	}

	h.logger.Debug(
		"XDS Service Changed",
		zap.String("xds_service_namespace", svc.GetNamespace()),
		zap.String("xds_service_name", svc.GetName()),
	)

	h.watches.notifyChanged(
		ctx,
		resourceRef{
			typeURL:      resourcesv3.ListenerType,
			resourceName: listenerName(svc.GetNamespace(), svc.GetName()),
		},
	)

	for _, cl := range svc.Spec.Clusters {
		h.watches.notifyChanged(
			ctx,
			resourceRef{
				typeURL:      resourcesv3.ClusterType,
				resourceName: clusterName(svc.GetNamespace(), svc.GetName(), cl.Name),
			},
		)

		h.watches.notifyChanged(
			ctx,
			resourceRef{
				typeURL:      resourcesv3.EndpointType,
				resourceName: clusterName(svc.GetNamespace(), svc.GetName(), cl.Name),
			},
		)
	}

	return nil
}

type endpointSliceChangedHandler struct {
	watches *watches
	logger  *zap.Logger

	servicesLister kxdslisters.XDSServiceLister
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

	svcs, err := h.servicesLister.List(labels.Everything())
	if err != nil {
		h.logger.Error("Could not list XDS services", zap.Error(err))
		return err
	}

	// O(n) accross all services isn't good. Yet that's the price of maintaining cross namespace localities.
	// Dropping this feature would allow us to narrow down the list of services to lookup by namespace.
	for _, svc := range svcs {
		for _, cl := range svc.Spec.Clusters {
			if matchesCluster(objMeta, svc, cl) {
				h.logger.Debug(
					"Endpoint changed",
					zap.String("xds_service_namespace", svc.GetNamespace()),
					zap.String("xds_service_name", svc.GetName()),
					zap.String("endpoint_name", objMeta.GetName()),
					zap.String("endpoint_namespace", objMeta.GetNamespace()),
				)

				h.watches.notifyChanged(
					ctx,
					resourceRef{
						typeURL: resourcesv3.EndpointType,
						resourceName: clusterName(
							svc.GetNamespace(),
							svc.GetName(),
							cl.Name,
						),
					},
				)
			}
		}
	}

	return nil
}

func matchesCluster(epSlice metav1.Object, xdsSvc *kxdsv1alpha1.XDSService, cluster kxdsv1alpha1.Cluster) bool {
	switch {
	case cluster.Service != nil:
		return matchesService(epSlice, xdsSvc, cluster.Service)
	case len(cluster.Localities) > 0:
		for _, loc := range cluster.Localities {
			if matchesService(epSlice, xdsSvc, loc.Service) {
				return true
			}
		}

		return false
	default:
		return false
	}
}

func matchesService(epSlice metav1.Object, xdsSvc *kxdsv1alpha1.XDSService, svc *kxdsv1alpha1.ServiceRef) bool {
	// If the name doesn't match then we're out.
	if svcName := epSlice.GetLabels()["kubernetes.io/service-name"]; svcName != svc.Name {
		return false
	}

	// If we don't have a specific namespace, then we match looking at the XDS service namespace.
	return xdsSvc.Namespace == epSlice.GetNamespace() || (svc.Namespace != "" && svc.Namespace == epSlice.GetNamespace())
}
