package gtc

import (
	"fmt"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	resourcesv3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	anyv1 "github.com/golang/protobuf/ptypes/any"
	gtcv1alpha1 "github.com/jlevesy/grpc-traffic-controller/api/gtc/v1alpha1"
	gtclisters "github.com/jlevesy/grpc-traffic-controller/client/listers/gtc/v1alpha1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type clusterHandler struct {
	grpcListeners gtclisters.GRPCListenerLister
}

func (h *clusterHandler) resolveResource(req resolveRequest) (*resolveResponse, error) {
	var (
		err      error
		versions = make([]string, len(req.resourceNames))
		response = resolveResponse{
			typeURL:   resourcesv3.ClusterType,
			resources: make([]*anyv1.Any, len(req.resourceNames)),
		}
	)

	for i, resourceName := range req.resourceNames {
		ref, err := parseXDSResourceName(resourceName)
		if err != nil {
			return nil, err
		}

		listener, err := h.grpcListeners.GRPCListeners(ref.Namespace).Get(ref.ListenerName)
		if err != nil {
			return nil, err
		}

		cl, err := extractClusterSpec(ref.ResourceName, listener)
		if err != nil {
			return nil, err
		}

		response.resources[i], err = encodeResource(req.typeUrl, makeCluster(resourceName, cl))
		if err != nil {
			return nil, err
		}

		versions[i] = listener.ResourceVersion
	}

	response.versionInfo, err = computeVersionInfo(versions)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func makeCluster(clusterName string, spec gtcv1alpha1.Cluster) *cluster.Cluster {
	c := cluster.Cluster{
		Name:                 clusterName,
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
		EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
			EdsConfig: &core.ConfigSource{
				ConfigSourceSpecifier: &core.ConfigSource_Ads{
					Ads: &core.AggregatedConfigSource{},
				},
			},
			ServiceName: clusterName,
		},
		LbPolicy: cluster.Cluster_ROUND_ROBIN,
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

	return &c
}

func extractClusterSpec(clusterName string, listener *gtcv1alpha1.GRPCListener) (gtcv1alpha1.Cluster, error) {
	if listener.Spec.DefaultCluster != nil {
		return gtcv1alpha1.Cluster{
			Name:        "default",
			MaxRequests: listener.Spec.DefaultCluster.MaxRequests,
			Service:     listener.Spec.DefaultCluster.Service,
		}, nil
	}

	for _, cl := range listener.Spec.Clusters {
		if cl.Name == clusterName {
			return cl, nil
		}
	}

	return gtcv1alpha1.Cluster{}, &clusterNotFoundError{wantName: clusterName, listener: listener}
}

type clusterNotFoundError struct {
	wantName string
	listener *gtcv1alpha1.GRPCListener
}

func (c *clusterNotFoundError) Error() string {
	return fmt.Sprintf("cluster with name %q does not exist on GRPCListener %s/%s", c.wantName, c.listener.Namespace, c.listener.Name)
}
