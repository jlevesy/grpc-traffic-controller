package gtc

import (
	"errors"
	"fmt"
	"strings"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	resourcesv3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
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
	response := newResolveResponse(resourcesv3.EndpointType, len(req.resourceNames))

	for i, resourceName := range req.resourceNames {
		backendRef, err := parseBackendName(resourceName)
		if err != nil {
			return nil, err
		}

		listeners, err := h.grpcListeners.GRPCListeners(backendRef.Namespace).Get(backendRef.ListenerName)
		if err != nil {
			return nil, err
		}

		if err := response.useResourceVersion(listeners.ResourceVersion); err != nil {
			return nil, err
		}

		backend, err := findBackendSpec(backendRef, listeners)
		if err != nil {
			return nil, err
		}

		eps, slicesVersions, err := h.makeLoadAssignment(req.nodeInfo, backendRef, listeners, backend)
		if err != nil {
			return nil, err
		}

		response.resources[i], err = encodeResource(req.typeUrl, eps)
		if err != nil {
			return nil, err
		}

		for _, v := range slicesVersions {
			if err := response.useResourceVersion(v); err != nil {
				return nil, err
			}
		}
	}

	return response, nil
}

func (h *endpointHandler) makeLoadAssignment(node *v3.Node, backendRef parsedBackendName, listener *gtcv1alpha1.GRPCListener, backendSpec gtcv1alpha1.Backend) (*endpointv3.ClusterLoadAssignment, []string, error) {
	switch {
	case backendSpec.Service != nil:
		return h.makeServiceLoadAssignment(node, backendRef, listener, backendSpec)
	case len(backendSpec.Localities) > 0:
		return h.makeLocalitiesLoadAssignment(backendRef, listener, backendSpec)
	default:
		return nil, nil, errors.New("unsupported non k8s service locality")
	}
}

func (h *endpointHandler) makeLocalitiesLoadAssignment(backendRef parsedBackendName, listener *gtcv1alpha1.GRPCListener, clusterSpec gtcv1alpha1.Backend) (*endpointv3.ClusterLoadAssignment, []string, error) {
	var (
		result = endpointv3.ClusterLoadAssignment{
			ClusterName: backendRef.String(),
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

func (h *endpointHandler) makeServiceLoadAssignment(node *v3.Node, backendRef parsedBackendName, listener *gtcv1alpha1.GRPCListener, clusterSpec gtcv1alpha1.Backend) (*endpointv3.ClusterLoadAssignment, []string, error) {
	result := endpointv3.ClusterLoadAssignment{
		ClusterName: backendRef.String(),
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

	result.Endpoints, err = makeServiceEndpoints(node, *clusterSpec.Service, endpointSlices)
	if err != nil {
		return nil, nil, err
	}

	versions := make([]string, len(endpointSlices))

	for i, s := range endpointSlices {
		versions[i] = s.ResourceVersion
	}

	return &result, versions, nil
}

type endpointGroup struct {
	zone string

	endpoints []endpointData
}

type endpointData struct {
	port     uint32
	endpoint kdiscoveryv1.Endpoint
}

func makeServiceEndpoints(node *v3.Node, serviceRef gtcv1alpha1.ServiceRef, epSlices []*kdiscoveryv1.EndpointSlice) ([]*endpointv3.LocalityLbEndpoints, error) {
	var currentZone string

	if node.Locality != nil {
		currentZone = node.Locality.Zone
	}

	localityEndpointsByPrio := [2]*endpointGroup{}

	for _, epSlice := range epSlices {
		port, ok := lookupK8sPort(serviceRef.Port, epSlice.Ports)
		if !ok {
			return nil, errors.New("no desired port found on the k8s endpoint slice")
		}

		for _, ep := range epSlice.Endpoints {
			if !derefBool(ep.Conditions.Ready) {
				continue
			}

			priority := 1

			if ep.Hints != nil && containsZone(ep.Hints.ForZones, currentZone) {
				priority = 0
			}

			prioGroup := localityEndpointsByPrio[priority]
			if prioGroup == nil {
				prioGroup = &endpointGroup{}
			}

			// loose capturing of the zone for the endpoint group.
			// not great, not terrible, that's metadata anyway.
			if prioGroup.zone == "" && ep.Zone != nil {
				prioGroup.zone = *ep.Zone
			}

			prioGroup.endpoints = append(
				prioGroup.endpoints,
				endpointData{
					port:     port,
					endpoint: ep,
				},
			)

			localityEndpointsByPrio[priority] = prioGroup
		}
	}

	var (
		result []*endpointv3.LocalityLbEndpoints

		// We consider that the distribution by prio should be accounted for
		// if we have at least one endpoint that was hinted for the zone we're
		// currently running in. Which means that we have at least one endpoint
		// in the priority 0 endpointGroup.
		enableZonePrio = localityEndpointsByPrio[0] != nil && len(localityEndpointsByPrio[0].endpoints) > 0
	)

	for prio, prioGroup := range localityEndpointsByPrio {
		if prioGroup == nil || len(prioGroup.endpoints) == 0 {
			continue
		}

		// Force priority zero if we detect that no hints are presents on all concerned endpoints.
		var appliedPrio uint32

		if enableZonePrio {
			appliedPrio = uint32(prio)
		}

		localityLbEndpoints := endpointv3.LocalityLbEndpoints{
			Locality: &core.Locality{
				Zone: prioGroup.zone,
				SubZone: fmt.Sprintf(
					"%s-%d",
					serviceRef.Name,
					appliedPrio,
				),
			},
			LoadBalancingWeight: wrapperspb.UInt32(1),
			Priority:            appliedPrio,
		}

		for _, ep := range prioGroup.endpoints {
			localityLbEndpoints.LbEndpoints = append(
				localityLbEndpoints.LbEndpoints,
				makeLbEndpoints(ep.endpoint, ep.port)...,
			)
		}

		result = append(result, &localityLbEndpoints)
	}

	return result, nil
}

func makeFlatLocalityLbEndpoints(serviceRef gtcv1alpha1.ServiceRef, epSlices []*kdiscoveryv1.EndpointSlice, weight, priority uint32) (*endpointv3.LocalityLbEndpoints, error) {
	var xdsEndpoints []*endpointv3.LbEndpoint

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

	return &endpointv3.LocalityLbEndpoints{
		Locality:            &core.Locality{SubZone: serviceRef.Name},
		LoadBalancingWeight: wrapperspb.UInt32(weight),
		Priority:            priority,
		LbEndpoints:         xdsEndpoints,
	}, nil
}

func makeLbEndpoints(ep kdiscoveryv1.Endpoint, port uint32) []*endpointv3.LbEndpoint {
	var eps []*endpointv3.LbEndpoint

	for _, addr := range ep.Addresses {
		eps = append(
			eps,
			makeLbEndpoint(ep, addr, port),
		)
	}

	return eps
}

func makeLbEndpoint(ep kdiscoveryv1.Endpoint, addr string, port uint32) *endpointv3.LbEndpoint {
	return &endpointv3.LbEndpoint{
		HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
			Endpoint: &endpointv3.Endpoint{
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

func containsZone(zones []kdiscoveryv1.ForZone, currZone string) bool {
	for _, z := range zones {
		if strings.EqualFold(z.Name, currZone) {
			return true
		}
	}

	return false
}

func derefBool(v *bool) bool {
	if v == nil {
		return false
	}

	return *v
}
