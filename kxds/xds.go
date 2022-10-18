package kxds

import (
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	corev1 "k8s.io/api/core/v1"
)

func makeListener(listenerName string, routeName string) *listener.Listener {
	httpConnManager := &hcm.HttpConnectionManager{
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				RouteConfigName: routeName,
				ConfigSource: &core.ConfigSource{
					ResourceApiVersion:    core.ApiVersion_V3,
					ConfigSourceSpecifier: &core.ConfigSource_Ads{Ads: &core.AggregatedConfigSource{}},
				},
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name: wellknown.Router,
			ConfigType: &hcm.HttpFilter_TypedConfig{
				TypedConfig: mustAny(&router.Router{}),
			},
		}},
	}

	return &listener.Listener{
		Name: listenerName,
		ApiListener: &listener.ApiListener{
			ApiListener: mustAny(httpConnManager),
		},
	}
}

func makeRoute(listenerName, routeName, clusterName string) *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name:             routeName,
		ValidateClusters: &wrapperspb.BoolValue{Value: true},
		VirtualHosts: []*route.VirtualHost{{
			Name:    routeName + "-local-service",
			Domains: []string{listenerName},
			Routes: []*route.Route{{
				Match: &route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{
						Prefix: "/",
					},
				},
				Action: &route.Route_Route{
					Route: &route.RouteAction{
						ClusterSpecifier: &route.RouteAction_Cluster{
							Cluster: clusterName,
						},
					},
				},
			}},
		}},
	}
}

func makeCluster(clusterName string) *cluster.Cluster {
	return &cluster.Cluster{
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
}

func makeLoadAssignment(clusterName string, destinationPort int, endpoints corev1.Endpoints) *endpoint.ClusterLoadAssignment {
	var xdsEndpoints []*endpoint.LbEndpoint

	for _, ep := range endpoints.Subsets {
		xdsEndpoints = append(
			xdsEndpoints,
			makeEndpointsFromSubset(ep, destinationPort)...,
		)
	}

	return &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{
			{
				Locality:            &core.Locality{SubZone: clusterName + "-k8s"},
				LoadBalancingWeight: &wrapperspb.UInt32Value{Value: 1},
				Priority:            0,
				LbEndpoints:         xdsEndpoints,
			},
		},
	}
}

func makeEndpointsFromSubset(ep corev1.EndpointSubset, port int) []*endpoint.LbEndpoint {
	eps := make([]*endpoint.LbEndpoint, len(ep.Addresses))

	for i, ep := range ep.Addresses {
		eps[i] = &endpoint.LbEndpoint{
			HostIdentifier: &endpoint.LbEndpoint_Endpoint{
				Endpoint: &endpoint.Endpoint{
					Address: &core.Address{
						Address: &core.Address_SocketAddress{
							SocketAddress: &core.SocketAddress{
								Protocol: core.SocketAddress_TCP,
								Address:  ep.IP,
								PortSpecifier: &core.SocketAddress_PortValue{
									PortValue: uint32(port),
								},
							},
						},
					},
				},
			},
		}

	}

	return eps
}

func mustAny(msg protoreflect.ProtoMessage) *anypb.Any {
	p, err := anypb.New(msg)
	if err != nil {
		panic(err)
	}

	return p
}
