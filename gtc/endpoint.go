package gtc

import (
	"errors"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	resourcesv3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	anyv1 "github.com/golang/protobuf/ptypes/any"
	gtcv1alpha1 "github.com/jlevesy/grpc-traffic-controller/api/gtc/v1alpha1"
	gtclisters "github.com/jlevesy/grpc-traffic-controller/client/listers/gtc/v1alpha1"
	"google.golang.org/protobuf/types/known/wrapperspb"
	kdiscoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	discoveryv1listers "k8s.io/client-go/listers/discovery/v1"
)

type endpointHandler struct {
	grpcListeners  gtclisters.GRPCListenerLister
	endpointSlices discoveryv1listers.EndpointSliceLister
}

func (h *endpointHandler) resolveResource(req resolveRequest) (*resolveResponse, error) {
	var (
		versions []string
		response = resolveResponse{
			typeURL:   resourcesv3.EndpointType,
			resources: make([]*anyv1.Any, len(req.resourceNames)),
		}
	)
	for i, resourceName := range req.resourceNames {
		ref, err := parseXDSResourceName(resourceName)
		if err != nil {
			return nil, err
		}

		listeners, err := h.grpcListeners.GRPCListeners(ref.Namespace).Get(ref.ListenerName)
		if err != nil {
			return nil, err
		}

		versions = append(versions, listeners.ResourceVersion)

		cl, err := extractClusterSpec(ref.ResourceName, listeners)
		if err != nil {
			return nil, err
		}

		eps, slicesVersions, err := h.makeLoadAssignment(listeners, cl)
		if err != nil {
			return nil, err
		}

		versions = append(versions, slicesVersions...)

		response.resources[i], err = encodeResource(req.typeUrl, eps)
		if err != nil {
			return nil, err
		}
	}

	var err error

	response.versionInfo, err = computeVersionInfo(versions)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (h *endpointHandler) makeLoadAssignment(listener *gtcv1alpha1.GRPCListener, clusterSpec gtcv1alpha1.Cluster) (*endpointv3.ClusterLoadAssignment, []string, error) {
	switch {
	case clusterSpec.Service != nil:
		return h.makeServiceLoadAssignment(listener, clusterSpec)
	case len(clusterSpec.Localities) > 0:
		return h.makeLocalitiesLoadAssignment(listener, clusterSpec)
	default:
		return nil, nil, errors.New("unsupported non k8s service locality")
	}
}

func (h *endpointHandler) makeLocalitiesLoadAssignment(listener *gtcv1alpha1.GRPCListener, clusterSpec gtcv1alpha1.Cluster) (*endpointv3.ClusterLoadAssignment, []string, error) {
	var (
		result = endpoint.ClusterLoadAssignment{
			ClusterName: clusterName(listener.Namespace, listener.Name, clusterSpec.Name),
			Endpoints:   make([]*endpointv3.LocalityLbEndpoints, len(clusterSpec.Localities)),
		}

		versions []string
	)

	for i, loc := range clusterSpec.Localities {
		ns := loc.Service.Namespace
		if ns == "" {
			ns = listener.Namespace
		}

		req, err := labels.NewRequirement(
			"kubernetes.io/service-name",
			selection.Equals,
			[]string{loc.Service.Name},
		)
		if err != nil {
			return nil, nil, err
		}

		endpointSlices, err := h.endpointSlices.EndpointSlices(ns).List(
			labels.NewSelector().Add(*req),
		)
		if err != nil {
			return nil, nil, err
		}

		result.Endpoints[i], err = makeFlatLocalityLbEndpoints(*loc.Service, endpointSlices, loc.Weight, loc.Priority)
		if err != nil {
			return nil, nil, err
		}

		for _, s := range endpointSlices {
			versions = append(versions, s.ResourceVersion)
		}
	}

	return &result, versions, nil
}

func (h *endpointHandler) makeServiceLoadAssignment(listener *gtcv1alpha1.GRPCListener, clusterSpec gtcv1alpha1.Cluster) (*endpointv3.ClusterLoadAssignment, []string, error) {
	result := endpoint.ClusterLoadAssignment{
		ClusterName: clusterName(listener.Namespace, listener.Name, clusterSpec.Name),
	}

	ns := clusterSpec.Service.Namespace
	if ns == "" {
		ns = listener.Namespace
	}

	req, err := labels.NewRequirement(
		"kubernetes.io/service-name",
		selection.Equals,
		[]string{clusterSpec.Service.Name},
	)
	if err != nil {
		return nil, nil, err
	}

	endpointSlices, err := h.endpointSlices.EndpointSlices(ns).List(
		labels.NewSelector().Add(*req),
	)
	if err != nil {
		return nil, nil, err
	}

	localityEndpoints, err := makeFlatLocalityLbEndpoints(*clusterSpec.Service, endpointSlices, 1, 0)
	if err != nil {
		return nil, nil, err
	}

	result.Endpoints = []*endpoint.LocalityLbEndpoints{localityEndpoints}

	versions := make([]string, len(endpointSlices))

	for i, s := range endpointSlices {
		versions[i] = s.ResourceVersion
	}

	return &result, versions, nil
}

func makeFlatLocalityLbEndpoints(serviceRef gtcv1alpha1.ServiceRef, epSlices []*kdiscoveryv1.EndpointSlice, weight, priority uint32) (*endpoint.LocalityLbEndpoints, error) {
	var xdsEndpoints []*endpoint.LbEndpoint

	for _, epSlice := range epSlices {
		port, ok := lookupK8sPort(serviceRef.Port, epSlice.Ports)
		if !ok {
			return nil, errors.New("no desired port found on the k8s endpoint slice")
		}

		for _, ep := range epSlice.Endpoints {
			if !derefBool(ep.Conditions.Ready) {
				continue
			}

			xdsEndpoints = append(
				xdsEndpoints,
				makeLbEndpoints(ep, port)...,
			)
		}

	}

	return &endpoint.LocalityLbEndpoints{
		Locality:            &core.Locality{SubZone: serviceRef.Name},
		LoadBalancingWeight: wrapperspb.UInt32(weight),
		Priority:            priority,
		LbEndpoints:         xdsEndpoints,
	}, nil
}

func makeLbEndpoints(ep kdiscoveryv1.Endpoint, port uint32) []*endpoint.LbEndpoint {
	var eps []*endpoint.LbEndpoint

	for _, addr := range ep.Addresses {
		eps = append(
			eps,
			makeLbEndpoint(ep, addr, port),
		)
	}

	return eps
}

func makeLbEndpoint(ep kdiscoveryv1.Endpoint, addr string, port uint32) *endpoint.LbEndpoint {
	return &endpoint.LbEndpoint{
		HostIdentifier: &endpoint.LbEndpoint_Endpoint{
			Endpoint: &endpoint.Endpoint{
				Address: &core.Address{
					Address: &core.Address_SocketAddress{
						SocketAddress: &core.SocketAddress{
							Protocol: core.SocketAddress_TCP,
							Address:  addr,
							PortSpecifier: &core.SocketAddress_PortValue{
								PortValue: port,
							},
						},
					},
				},
			},
		},
	}
}

func lookupK8sPort(k8sSvc gtcv1alpha1.PortRef, epPorts []kdiscoveryv1.EndpointPort) (uint32, bool) {
	if k8sSvc.Name != "" {
		for _, p := range epPorts {
			if p.Name != nil && p.Port != nil && *p.Name == k8sSvc.Name {
				return uint32(*p.Port), true
			}

		}

		return 0, false
	}

	for _, p := range epPorts {
		if p.Name != nil && p.Port != nil && *p.Port == k8sSvc.Number {
			return uint32(*p.Port), true
		}
	}

	return 0, false
}

func derefBool(v *bool) bool {
	if v == nil {
		return false
	}

	return *v
}
