package kxds

import (
	"fmt"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	resourcesv3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	anyv1 "github.com/golang/protobuf/ptypes/any"
	kxdsv1alpha1 "github.com/jlevesy/grpc-traffic-controller/api/kxds/v1alpha1"
	kxdslisters "github.com/jlevesy/grpc-traffic-controller/client/listers/kxds/v1alpha1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type clusterHandler struct {
	xdsServices kxdslisters.XDSServiceLister
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

		svc, err := h.xdsServices.XDSServices(ref.Namespace).Get(ref.ServiceName)
		if err != nil {
			return nil, err
		}

		cl, err := extractClusterSpec(ref.ResourceName, svc)
		if err != nil {
			return nil, err
		}

		response.resources[i], err = encodeResource(req.typeUrl, makeCluster(resourceName, cl))
		if err != nil {
			return nil, err
		}

		versions[i] = svc.ResourceVersion
	}

	response.versionInfo, err = computeVersionInfo(versions)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func makeCluster(clusterName string, spec kxdsv1alpha1.Cluster) *cluster.Cluster {
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

func extractClusterSpec(clusterName string, xdsSvc *kxdsv1alpha1.XDSService) (kxdsv1alpha1.Cluster, error) {
	if xdsSvc.Spec.DefaultCluster != nil {
		return kxdsv1alpha1.Cluster{
			Name:        "default",
			MaxRequests: xdsSvc.Spec.DefaultCluster.MaxRequests,
			Service:     xdsSvc.Spec.DefaultCluster.Service,
		}, nil
	}

	for _, cl := range xdsSvc.Spec.Clusters {
		if cl.Name == clusterName {
			return cl, nil
		}
	}

	return kxdsv1alpha1.Cluster{}, &clusterNotFoundError{wantName: clusterName, svc: xdsSvc}
}

type clusterNotFoundError struct {
	wantName string
	svc      *kxdsv1alpha1.XDSService
}

func (c *clusterNotFoundError) Error() string {
	return fmt.Sprintf("cluster with name %q does not exist on XDS service %s/%s", c.wantName, c.svc.Namespace, c.svc.Name)
}
