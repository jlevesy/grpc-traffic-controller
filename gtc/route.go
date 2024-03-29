package gtc

import (
	"errors"
	"fmt"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	gtcv1alpha1 "github.com/jlevesy/grpc-traffic-controller/api/gtc/v1alpha1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func makeRouteConfig(listenerName string, listener *gtcv1alpha1.GRPCListener) (*route.RouteConfiguration, error) {
	routes := make([]*route.Route, len(listener.Spec.Routes))

	for routeID, routeSpec := range listener.Spec.Routes {
		match, err := makeRouteMatch(routeSpec)
		if err != nil {
			return nil, err
		}

		filterOverrides, err := makeFilterOverrides(routeSpec.Interceptors)
		if err != nil {
			return nil, err
		}

		weighedClusters, err := makeWeightedClusters(
			listener.Namespace,
			listener.Name,
			routeID,
			routeSpec,
		)
		if err != nil {
			return nil, err
		}

		var routeRetryPolicy *route.RetryPolicy

		if routeSpec.Retry != nil {
			routeRetryPolicy = makeRetryPolicy(routeSpec.Retry)
		}

		hashPolicy, err := makeHashPolicy(routeSpec.HashPolicy)
		if err != nil {
			return nil, err
		}

		routes[routeID] = &route.Route{
			Match:                match,
			TypedPerFilterConfig: filterOverrides,
			Action: &route.Route_Route{
				Route: &route.RouteAction{
					HashPolicy:  hashPolicy,
					RetryPolicy: routeRetryPolicy,
					MaxStreamDuration: &route.RouteAction_MaxStreamDuration{
						MaxStreamDuration:    makeDuration(routeSpec.MaxStreamDuration),
						GrpcTimeoutHeaderMax: makeDuration(routeSpec.GrpcTimeoutHeaderMax),
					},
					ClusterSpecifier: &route.RouteAction_WeightedClusters{
						WeightedClusters: weighedClusters,
					},
				},
			},
		}
	}

	var listenerRetryPolicy *route.RetryPolicy

	if listener.Spec.Retry != nil {
		listenerRetryPolicy = makeRetryPolicy(listener.Spec.Retry)
	}

	return &route.RouteConfiguration{
		Name:             routeConfigName(listener.Namespace, listener.Name),
		ValidateClusters: &wrapperspb.BoolValue{Value: true},
		VirtualHosts: []*route.VirtualHost{
			{
				Name:        vHostName(listener.Namespace, listener.Name),
				Domains:     []string{listenerName},
				Routes:      routes,
				RetryPolicy: listenerRetryPolicy,
			},
		},
	}, nil
}

var matchAll = route.RouteMatch{
	PathSpecifier: &route.RouteMatch_Prefix{
		Prefix: "/",
	},
}

func makeRouteMatch(spec gtcv1alpha1.Route) (*route.RouteMatch, error) {
	if spec.Matcher == nil {
		return &matchAll, nil
	}

	var match route.RouteMatch

	switch {
	case spec.Matcher.Method != nil:
		match.PathSpecifier = &route.RouteMatch_Path{
			Path: spec.Matcher.Method.Path(),
		}
	case spec.Matcher.Service != nil:
		match.PathSpecifier = &route.RouteMatch_Prefix{
			Prefix: spec.Matcher.Service.Prefix(),
		}
	case spec.Matcher.Namespace != nil:
		match.PathSpecifier = &route.RouteMatch_Prefix{
			Prefix: "/" + *spec.Matcher.Namespace,
		}
	default:
		match.PathSpecifier = matchAll.PathSpecifier
	}

	if spec.Matcher.Fraction != nil {
		fraction, err := makeFractionalPercent(spec.Matcher.Fraction)
		if err != nil {
			return nil, err
		}

		match.RuntimeFraction = &corev3.RuntimeFractionalPercent{
			DefaultValue: fraction,
		}
	}

	match.Headers = make([]*route.HeaderMatcher, len(spec.Matcher.Metadata))

	for i, metadataMatcherSpec := range spec.Matcher.Metadata {
		var err error

		match.Headers[i], err = makeMetadataMatcher(metadataMatcherSpec)
		if err != nil {
			return nil, err
		}
	}

	return &match, nil
}

func makeMetadataMatcher(spec gtcv1alpha1.MetadataMatcher) (*route.HeaderMatcher, error) {
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

func makeRegexMatcher(spec *gtcv1alpha1.RegexMatcher) (*matcher.RegexMatcher, error) {
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

func makeWeightedClusters(namespace, name string, routeID int, routeSpec gtcv1alpha1.Route) (*route.WeightedCluster, error) {
	var (
		totalWeight     uint32
		weighedClusters = make([]*route.WeightedCluster_ClusterWeight, len(routeSpec.Backends))
	)

	for backendID, backend := range routeSpec.Backends {
		filterOverrides, err := makeFilterOverrides(backend.Interceptors)
		if err != nil {
			return nil, err
		}

		totalWeight += backend.Weight
		weighedClusters[backendID] = &route.WeightedCluster_ClusterWeight{
			Name:                 backendName(namespace, name, routeID, backendID),
			Weight:               wrapperspb.UInt32(backend.Weight),
			TypedPerFilterConfig: filterOverrides,
		}
	}

	return &route.WeightedCluster{
		TotalWeight: wrapperspb.UInt32(totalWeight),
		Clusters:    weighedClusters,
	}, nil
}

func makeRetryPolicy(spec *gtcv1alpha1.RetryPolicy) *route.RetryPolicy {
	var (
		numRetries uint32 = 1
		backoff    *route.RetryPolicy_RetryBackOff
	)

	if spec.NumRetries != nil {
		numRetries = *spec.NumRetries
	}

	if spec.Backoff != nil {
		backoff = &route.RetryPolicy_RetryBackOff{
			BaseInterval: durationpb.New(spec.Backoff.BaseInterval.Duration),
		}

		if spec.Backoff.MaxInterval != nil {
			backoff.MaxInterval = durationpb.New(spec.Backoff.MaxInterval.Duration)
		}
	}

	return &route.RetryPolicy{
		RetryOn:      strings.Join(spec.RetryOn, ","),
		NumRetries:   wrapperspb.UInt32(numRetries),
		RetryBackOff: backoff,
	}
}

func makeHashPolicy(policies []gtcv1alpha1.HashPolicy) ([]*route.RouteAction_HashPolicy, error) {
	if len(policies) == 0 {
		return nil, nil
	}

	result := make([]*route.RouteAction_HashPolicy, len(policies))

	for i, policy := range policies {
		switch {
		case policy.Metadata != "":
			result[i] = &route.RouteAction_HashPolicy{
				PolicySpecifier: &route.RouteAction_HashPolicy_Header_{
					Header: &route.RouteAction_HashPolicy_Header{
						HeaderName: policy.Metadata,
					},
				},
				Terminal: policy.Terminal,
			}
		case policy.Channel != nil && *policy.Channel:
			result[i] = &route.RouteAction_HashPolicy{
				PolicySpecifier: &route.RouteAction_HashPolicy_FilterState_{
					FilterState: &route.RouteAction_HashPolicy_FilterState{
						Key: "io.grpc.channel_id",
					},
				},
				Terminal: policy.Terminal,
			}

		default:
			return nil, errors.New("malformed hash policy")
		}
	}

	return result, nil
}
