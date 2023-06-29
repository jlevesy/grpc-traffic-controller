package kxds

import (
	"errors"
	"fmt"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	kxdsv1alpha1 "github.com/jlevesy/kxds/api/kxds/v1alpha1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func makeRouteConfig(listenerName string, svc *kxdsv1alpha1.XDSService) (*route.RouteConfiguration, error) {
	var (
		routeSpecs = svc.Spec.Routes
		routesLen  = len(svc.Spec.Routes)
	)

	// If the service has no route, then populate a default one that points to the default clusterRef.
	if len(svc.Spec.Routes) == 0 {
		routesLen = 1
		routeSpecs = []kxdsv1alpha1.Route{
			{
				Clusters: []kxdsv1alpha1.ClusterRef{
					{Name: "default", Weight: 1},
				},
			},
		}
	}

	routes := make([]*route.Route, routesLen)

	for i, routeSpec := range routeSpecs {
		match, err := makeRouteMatch(routeSpec)
		if err != nil {
			return nil, err
		}

		routes[i] = &route.Route{
			Match: match,
			Action: &route.Route_Route{
				Route: &route.RouteAction{
					MaxStreamDuration: &route.RouteAction_MaxStreamDuration{
						MaxStreamDuration:    makeDuration(routeSpec.MaxStreamDuration),
						GrpcTimeoutHeaderMax: makeDuration(routeSpec.GrpcTimeoutHeaderMax),
					},
					ClusterSpecifier: &route.RouteAction_WeightedClusters{
						WeightedClusters: makeWeightedClusters(svc.Namespace, svc.Name, routeSpec),
					},
				},
			},
		}
	}

	return &route.RouteConfiguration{
		Name:             routeConfigName(svc.Namespace, svc.Name),
		ValidateClusters: &wrapperspb.BoolValue{Value: true},
		VirtualHosts: []*route.VirtualHost{
			{
				Name:    vHostName(svc.Namespace, svc.Name),
				Domains: []string{listenerName},
				Routes:  routes,
			},
		},
	}, nil
}

func makeRouteMatch(spec kxdsv1alpha1.Route) (*route.RouteMatch, error) {
	var match route.RouteMatch

	switch {
	case spec.Path.Regex != nil:
		regexMatcher, err := makeRegexMatcher(spec.Path.Regex)
		if err != nil {
			return nil, err
		}

		match.PathSpecifier = &route.RouteMatch_SafeRegex{
			SafeRegex: regexMatcher,
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

	match.CaseSensitive = wrapperspb.Bool(spec.CaseSensitive)
	match.Headers = make([]*route.HeaderMatcher, len(spec.Headers))

	if spec.RuntimeFraction != nil {
		fraction, err := makeFractionalPercent(spec.RuntimeFraction)
		if err != nil {
			return nil, err
		}

		match.RuntimeFraction = &core.RuntimeFractionalPercent{
			DefaultValue: fraction,
		}
	}

	for i, headerMatcherSpec := range spec.Headers {
		var err error

		match.Headers[i], err = makeHeaderMatcher(headerMatcherSpec)
		if err != nil {
			return nil, err
		}
	}

	return &match, nil
}

func makeHeaderMatcher(spec kxdsv1alpha1.HeaderMatcher) (*route.HeaderMatcher, error) {
	matcher := route.HeaderMatcher{
		Name:        spec.Name,
		InvertMatch: spec.Invert,
	}

	switch {
	case spec.Exact != nil:
		matcher.HeaderMatchSpecifier = &route.HeaderMatcher_ExactMatch{
			ExactMatch: *spec.Exact,
		}
	case spec.Regex != nil:
		regexMatcher, err := makeRegexMatcher(spec.Regex)
		if err != nil {
			return nil, err
		}

		matcher.HeaderMatchSpecifier = &route.HeaderMatcher_SafeRegexMatch{
			SafeRegexMatch: regexMatcher,
		}
	case spec.Range != nil:
		matcher.HeaderMatchSpecifier = &route.HeaderMatcher_RangeMatch{
			RangeMatch: &typev3.Int64Range{
				Start: spec.Range.Start,
				End:   spec.Range.End,
			},
		}
	case spec.Present != nil:
		matcher.HeaderMatchSpecifier = &route.HeaderMatcher_PresentMatch{
			PresentMatch: *spec.Present,
		}
	case spec.Prefix != nil:
		matcher.HeaderMatchSpecifier = &route.HeaderMatcher_PrefixMatch{
			PrefixMatch: *spec.Prefix,
		}
	case spec.Suffix != nil:
		matcher.HeaderMatchSpecifier = &route.HeaderMatcher_SuffixMatch{
			SuffixMatch: *spec.Suffix,
		}
	default:
		return nil, errors.New("invalid header matcher")

	}

	return &matcher, nil
}

func makeRegexMatcher(spec *kxdsv1alpha1.RegexMatcher) (*matcher.RegexMatcher, error) {
	if spec.Engine != "re2" {
		return nil, fmt.Errorf("unsupported engine %q", spec.Engine)
	}

	if spec.Regex == "" {
		return nil, errors.New("blank regex")
	}

	return &matcher.RegexMatcher{
		Regex: spec.Regex,
		EngineType: &matcher.RegexMatcher_GoogleRe2{
			GoogleRe2: &matcher.RegexMatcher_GoogleRE2{},
		},
	}, nil
}

func makeWeightedClusters(namespace, name string, routeSpec kxdsv1alpha1.Route) *route.WeightedCluster {
	var (
		totalWeight     uint32
		weighedClusters = make([]*route.WeightedCluster_ClusterWeight, len(routeSpec.Clusters))
	)

	for i, clusterRef := range routeSpec.Clusters {
		totalWeight += clusterRef.Weight
		weighedClusters[i] = &route.WeightedCluster_ClusterWeight{
			Name:   clusterName(namespace, name, clusterRef.Name),
			Weight: wrapperspb.UInt32(clusterRef.Weight),
		}
	}

	return &route.WeightedCluster{
		TotalWeight: wrapperspb.UInt32(totalWeight),
		Clusters:    weighedClusters,
	}
}
