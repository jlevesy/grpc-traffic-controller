package kxds

import (
	"errors"
	"fmt"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	corev1 "k8s.io/api/core/v1"
	ktypes "k8s.io/apimachinery/pkg/types"

	kxdsv1alpha1 "github.com/jlevesy/kxds/api/v1alpha1"
)

type xdsService struct {
	listener        types.Resource
	routeConfig     types.Resource
	clusters        []types.Resource
	loadAssignments []types.Resource
}

func makeXDSService(svc kxdsv1alpha1.XDSService, k8sEndpoints map[ktypes.NamespacedName]corev1.Endpoints) (xdsService, error) {
	var (
		err error

		resourcePrefix  = "kxds" + "." + svc.Name + "." + svc.Namespace + "."
		listenerName    = svc.Spec.Listener
		routeConfigName = resourcePrefix + "routeconfig"

		xdsSvc = xdsService{
			listener: makeListener(listenerName, routeConfigName),
			clusters: make([]types.Resource, len(svc.Spec.Clusters)),
		}
	)

	xdsSvc.routeConfig, err = makeRouteConfig(resourcePrefix, routeConfigName, listenerName, svc.Spec.Routes)
	if err != nil {
		return xdsSvc, err
	}

	for i, clusterSpec := range svc.Spec.Clusters {
		clusterName := resourcePrefix + clusterSpec.Name

		xdsSvc.clusters[i] = makeCluster(clusterName)

		loadAssignment, err := makeLoadAssignment(
			clusterName,
			svc.Namespace,
			clusterSpec.Localities,
			k8sEndpoints,
		)
		if err != nil {
			return xdsSvc, err
		}

		xdsSvc.loadAssignments = append(xdsSvc.loadAssignments, loadAssignment)
	}

	return xdsSvc, nil
}

func makeListener(listenerName string, routeConfigName string) *listener.Listener {
	httpConnManager := &hcm.HttpConnectionManager{
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				RouteConfigName: routeConfigName,
				ConfigSource: &core.ConfigSource{
					ResourceApiVersion:    core.ApiVersion_V3,
					ConfigSourceSpecifier: &core.ConfigSource_Ads{Ads: &core.AggregatedConfigSource{}},
				},
			},
		},
		HttpFilters: []*hcm.HttpFilter{
			{
				Name: wellknown.Router,
				ConfigType: &hcm.HttpFilter_TypedConfig{
					TypedConfig: mustAny(&router.Router{}),
				},
			},
		},
	}

	return &listener.Listener{
		Name: listenerName,
		ApiListener: &listener.ApiListener{
			ApiListener: mustAny(httpConnManager),
		},
	}
}

func makeRouteConfig(resourcePrefix, routeConfigName, listenerName string, routeSpecs []kxdsv1alpha1.Route) (*route.RouteConfiguration, error) {
	routes := make([]*route.Route, len(routeSpecs))

	for i, routeSpec := range routeSpecs {
		match, err := makeRouteMatch(routeSpec)
		if err != nil {
			return nil, err
		}

		routes[i] = &route.Route{
			Match: match,
			Action: &route.Route_Route{
				Route: &route.RouteAction{
					ClusterSpecifier: &route.RouteAction_WeightedClusters{
						WeightedClusters: makeWeightedClusters(resourcePrefix, routeSpec),
					},
				},
			},
		}
	}

	return &route.RouteConfiguration{
		Name:             routeConfigName,
		ValidateClusters: &wrapperspb.BoolValue{Value: true},
		VirtualHosts: []*route.VirtualHost{
			{
				Name:    resourcePrefix + "vhost",
				Domains: []string{listenerName},
				Routes:  routes,
			},
		},
	}, nil
}

func makeRouteMatch(spec kxdsv1alpha1.Route) (*route.RouteMatch, error) {
	var match route.RouteMatch

	switch {
	case spec.Path.Regex.Regex != "":
		if spec.Path.Regex.Engine != "re2" {
			return nil, fmt.Errorf("unsupported engine %q", spec.Path.Regex.Engine)
		}

		match.PathSpecifier = &route.RouteMatch_SafeRegex{
			SafeRegex: &matcher.RegexMatcher{
				Regex: spec.Path.Regex.Regex,
				EngineType: &matcher.RegexMatcher_GoogleRe2{
					GoogleRe2: &matcher.RegexMatcher_GoogleRE2{},
				},
			},
		}

	case spec.Path.Path != "":
		match.PathSpecifier = &route.RouteMatch_Path{
			Path: spec.Path.Path,
		}
	default:
		match.PathSpecifier = &route.RouteMatch_Prefix{
			Prefix: spec.Path.Prefix,
		}
	}

	return &match, nil
}

func makeWeightedClusters(resourcePrefix string, routeSpec kxdsv1alpha1.Route) *route.WeightedCluster {
	var (
		totalWeight     uint32
		weighedClusters = make([]*route.WeightedCluster_ClusterWeight, len(routeSpec.Clusters))
	)

	for i, clusterRef := range routeSpec.Clusters {
		totalWeight += clusterRef.Weight
		weighedClusters[i] = &route.WeightedCluster_ClusterWeight{
			Name:   resourcePrefix + clusterRef.Name,
			Weight: wrapperspb.UInt32(clusterRef.Weight),
		}
	}

	return &route.WeightedCluster{
		TotalWeight: wrapperspb.UInt32(totalWeight),
		Clusters:    weighedClusters,
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

func makeLoadAssignment(clusterName, namespace string, localities []kxdsv1alpha1.Locality, k8sEndpoints map[ktypes.NamespacedName]corev1.Endpoints) (*endpoint.ClusterLoadAssignment, error) {
	xdsLocalities := make([]*endpoint.LocalityLbEndpoints, len(localities))

	for i, locSpec := range localities {
		if locSpec.Service == nil {
			return nil, errors.New("unsupported non k8s service locality")
		}

		k8sEndpoint, ok := k8sEndpoints[ktypes.NamespacedName{Namespace: namespace, Name: locSpec.Service.Name}]
		if !ok {
			return nil, errors.New("no k8s endpoints found")
		}

		var err error

		xdsLocalities[i], err = makeK8sLocality(locSpec, k8sEndpoint)
		if err != nil {
			return nil, fmt.Errorf("could not build cluster %q: %w", clusterName, err)
		}

	}

	return &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints:   xdsLocalities,
	}, nil
}

func makeK8sLocality(locSpec kxdsv1alpha1.Locality, k8sEndpoint corev1.Endpoints) (*endpoint.LocalityLbEndpoints, error) {
	var xdsEndpoints []*endpoint.LbEndpoint

	for _, ep := range k8sEndpoint.Subsets {
		port, ok := lookupK8sPort(locSpec.Service.Port, ep)
		if !ok {
			return nil, errors.New("no desired port found on the k8s endpoint")
		}

		xdsEndpoints = append(
			xdsEndpoints,
			makeEndpointsFromSubset(ep, port)...,
		)
	}

	return &endpoint.LocalityLbEndpoints{
		Locality:            &core.Locality{SubZone: locSpec.Service.Name},
		LoadBalancingWeight: wrapperspb.UInt32(locSpec.Weight),
		Priority:            locSpec.Priority,
		LbEndpoints:         xdsEndpoints,
	}, nil

}

func makeEndpointsFromSubset(ep corev1.EndpointSubset, port uint32) []*endpoint.LbEndpoint {
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
									PortValue: port,
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

func lookupK8sPort(k8sSvc kxdsv1alpha1.K8sPort, eps corev1.EndpointSubset) (uint32, bool) {
	if k8sSvc.Name != "" {
		for _, p := range eps.Ports {
			if p.Name == k8sSvc.Name {
				return uint32(p.Port), true
			}

		}

		return 0, false
	}

	for _, p := range eps.Ports {
		if p.Port == k8sSvc.Number {
			return uint32(p.Port), true
		}
	}

	return 0, false
}

func mustAny(msg protoreflect.ProtoMessage) *anypb.Any {
	p, err := anypb.New(msg)
	if err != nil {
		panic(err)
	}

	return p
}
