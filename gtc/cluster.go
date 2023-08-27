package gtc

import (
	"fmt"
	"strings"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	resourcesv3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	gtcv1alpha1 "github.com/jlevesy/grpc-traffic-controller/api/gtc/v1alpha1"
	gtclisters "github.com/jlevesy/grpc-traffic-controller/client/listers/gtc/v1alpha1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type clusterHandler struct {
	grpcListeners gtclisters.GRPCListenerLister
}

func (h *clusterHandler) resolveResource(req resolveRequest) (*resolveResponse, error) {
	response := newResolveResponse(resourcesv3.ClusterType, len(req.resourceNames))

	for i, resourceName := range req.resourceNames {
		backendRef, err := parseBackendName(resourceName)
		if err != nil {
			return nil, err
		}

		listener, err := h.grpcListeners.GRPCListeners(backendRef.Namespace).Get(backendRef.ListenerName)
		if err != nil {
			return nil, err
		}

		backend, err := findBackendSpec(backendRef, listener)
		if err != nil {
			return nil, err
		}

		response.resources[i], err = encodeResource(
			req.typeUrl,
			makeCluster(resourceName, backend),
		)
		if err != nil {
			return nil, err
		}

		if err := response.useResourceVersion(listener.ResourceVersion); err != nil {
			return nil, err
		}
	}

	return response, nil
}

func makeCluster(clusterName string, spec gtcv1alpha1.Backend) *cluster.Cluster {
	c := cluster.Cluster{
		Name:                 clusterName,
		LbPolicy:             makeLBPolicy(spec.LBPolicy),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
		EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
			EdsConfig: &core.ConfigSource{
				ConfigSourceSpecifier: &core.ConfigSource_Ads{
					Ads: &core.AggregatedConfigSource{},
				},
			},
			ServiceName: clusterName,
		},
	}

	// gRPC xDS only supports max requests, and will always look to the first value of the first threshold.
	if spec.MaxRequests != nil {
		c.CircuitBreakers = &cluster.CircuitBreakers{
			Thresholds: []*cluster.CircuitBreakers_Thresholds{
				{
					MaxRequests: wrapperspb.UInt32(*spec.MaxRequests),
				},
			},
		}
	}

	if spec.RingHashConfig != nil {
		c.LbConfig = &cluster.Cluster_RingHashLbConfig_{
			RingHashLbConfig: &cluster.Cluster_RingHashLbConfig{
				MinimumRingSize: wrapperspb.UInt64(spec.RingHashConfig.MinRingSize),
				MaximumRingSize: wrapperspb.UInt64(spec.RingHashConfig.MaxRingSize),
				HashFunction:    cluster.Cluster_RingHashLbConfig_XX_HASH,
			},
		}
	}

	return &c
}

var emptyBackend gtcv1alpha1.Backend

func findBackendSpec(backendRef parsedBackendName, listener *gtcv1alpha1.GRPCListener) (gtcv1alpha1.Backend, error) {
	if backendRef.RouteID > len(listener.Spec.Routes)-1 {
		return emptyBackend, &routeNotFoundError{
			wantRouteID: backendRef.RouteID,
			listener:    listener,
		}
	}

	route := listener.Spec.Routes[backendRef.RouteID]

	if backendRef.BackendID > len(route.Backends)-1 {
		return emptyBackend, &backendNotFoundError{
			routeID:       backendRef.RouteID,
			wantBackendID: backendRef.BackendID,
			listener:      listener,
		}
	}

	return route.Backends[backendRef.BackendID], nil
}

func makeLBPolicy(p string) cluster.Cluster_LbPolicy {
	switch strings.ToLower(p) {
	case "ringhash", "ring_hash":
		return cluster.Cluster_RING_HASH
	case "roundrobin", "round_robin":
		fallthrough
	default:
		return cluster.Cluster_ROUND_ROBIN

	}
}

type routeNotFoundError struct {
	wantRouteID int
	listener    *gtcv1alpha1.GRPCListener
}

func (c *routeNotFoundError) Error() string {
	return fmt.Sprintf(
		"route %d does not exist on GRPCListener %s/%s",
		c.wantRouteID,
		c.listener.Namespace,
		c.listener.Name,
	)
}

type backendNotFoundError struct {
	routeID       int
	wantBackendID int
	listener      *gtcv1alpha1.GRPCListener
}

func (c *backendNotFoundError) Error() string {
	return fmt.Sprintf(
		"backend with ID %d does not exist under the route %d of the GRPCListener %s/%s",
		c.wantBackendID,
		c.routeID,
		c.listener.Namespace,
		c.listener.Name,
	)
}
