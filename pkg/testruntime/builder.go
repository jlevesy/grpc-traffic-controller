package testruntime

import (
	"time"

	gtcv1alpha1 "github.com/jlevesy/grpc-traffic-controller/api/gtc/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Ptr[T any](v T) *T {
	return &v
}

func DurationPtr(d time.Duration) *metav1.Duration {
	return &metav1.Duration{
		Duration: d,
	}
}

type ClusterOption func(c *gtcv1alpha1.Cluster)

func WithMaxRequests(req uint32) ClusterOption {
	return func(c *gtcv1alpha1.Cluster) {
		c.MaxRequests = &req
	}
}

func WithServiceRef(s gtcv1alpha1.ServiceRef) ClusterOption {
	return func(c *gtcv1alpha1.Cluster) {
		c.Service = &s
	}
}

func WithLocalities(l ...gtcv1alpha1.Locality) ClusterOption {
	return func(c *gtcv1alpha1.Cluster) {
		c.Localities = l
	}
}

func BuildCluster(name string, opts ...ClusterOption) gtcv1alpha1.Cluster {
	c := gtcv1alpha1.Cluster{
		Name: name,
	}

	for _, opt := range opts {
		opt(&c)
	}

	return c
}

type LocalityOption func(l *gtcv1alpha1.Locality)

func WithLocalityWeight(weight uint32) LocalityOption {
	return func(l *gtcv1alpha1.Locality) {
		l.Weight = weight
	}
}

func WithLocalityPriority(priority uint32) LocalityOption {
	return func(l *gtcv1alpha1.Locality) {
		l.Priority = priority
	}
}

func WithLocalityServiceRef(s gtcv1alpha1.ServiceRef) LocalityOption {
	return func(l *gtcv1alpha1.Locality) {
		l.Service = &s
	}
}

func BuildLocality(opts ...LocalityOption) gtcv1alpha1.Locality {
	l := gtcv1alpha1.Locality{
		Weight: 1,
	}

	for _, opt := range opts {
		opt(&l)
	}

	return l
}

type DefaultClusterOption func(c *gtcv1alpha1.DefaultCluster)

func BuildDefaultCluster(opts ...DefaultClusterOption) gtcv1alpha1.DefaultCluster {
	var c gtcv1alpha1.DefaultCluster

	for _, o := range opts {
		o(&c)
	}

	return c
}

func WithDefaultServiceRef(s gtcv1alpha1.ServiceRef) DefaultClusterOption {
	return func(c *gtcv1alpha1.DefaultCluster) {
		c.Service = &s
	}
}

func HeaderInvertMatch(in gtcv1alpha1.HeaderMatcher) gtcv1alpha1.HeaderMatcher {
	in.Invert = true
	return in
}

func HeaderExactMatch(name, value string) gtcv1alpha1.HeaderMatcher {
	return gtcv1alpha1.HeaderMatcher{
		Name:  name,
		Exact: &value,
	}
}

func HeaderPresentMatch(name string, present bool) gtcv1alpha1.HeaderMatcher {
	return gtcv1alpha1.HeaderMatcher{
		Name:    name,
		Present: &present,
	}
}

func HeaderPrefixMatch(name, prefix string) gtcv1alpha1.HeaderMatcher {
	return gtcv1alpha1.HeaderMatcher{
		Name:   name,
		Prefix: &prefix,
	}
}

func HeaderSuffixMatch(name, suffix string) gtcv1alpha1.HeaderMatcher {
	return gtcv1alpha1.HeaderMatcher{
		Name:   name,
		Suffix: &suffix,
	}
}

type RouteOption func(r *gtcv1alpha1.Route)

func WithHeaderMatchers(matchers ...gtcv1alpha1.HeaderMatcher) RouteOption {
	return func(r *gtcv1alpha1.Route) {
		r.Headers = matchers
	}
}

func WithRouteMaxStreamDuration(d time.Duration) RouteOption {
	return func(r *gtcv1alpha1.Route) {
		r.MaxStreamDuration = &metav1.Duration{Duration: d}
	}
}

func WithRuntimeFraction(fr gtcv1alpha1.Fraction) RouteOption {
	return func(r *gtcv1alpha1.Route) {
		r.RuntimeFraction = &fr
	}
}

func WithClusterRefs(refs ...gtcv1alpha1.ClusterRef) RouteOption {
	return func(r *gtcv1alpha1.Route) {
		r.Clusters = refs
	}
}

func WithPathMatcher(pm gtcv1alpha1.PathMatcher) RouteOption {
	return func(r *gtcv1alpha1.Route) {
		r.Path = pm
	}
}

func WithCaseSensitive(v bool) RouteOption {
	return func(r *gtcv1alpha1.Route) {
		r.CaseSensitive = v
	}
}

func BuildRoute(opts ...RouteOption) gtcv1alpha1.Route {
	r := gtcv1alpha1.Route{
		Path: gtcv1alpha1.PathMatcher{
			Prefix: "/",
		},
	}

	for _, opt := range opts {
		opt(&r)
	}

	return r
}

func BuildSingleRoute(clusterName string) gtcv1alpha1.Route {
	return BuildRoute(
		WithClusterRefs(
			gtcv1alpha1.ClusterRef{
				Name:   clusterName,
				Weight: 1,
			},
		),
	)
}

type GRPCListenerOpt func(s *gtcv1alpha1.GRPCListener)

func WithFilters(fs ...gtcv1alpha1.Filter) GRPCListenerOpt {
	return func(s *gtcv1alpha1.GRPCListener) {
		s.Spec.Filters = fs
	}
}

func WithRoutes(rs ...gtcv1alpha1.Route) GRPCListenerOpt {
	return func(s *gtcv1alpha1.GRPCListener) {
		s.Spec.Routes = rs
	}
}

func WithClusters(cs ...gtcv1alpha1.Cluster) GRPCListenerOpt {
	return func(s *gtcv1alpha1.GRPCListener) {
		s.Spec.Clusters = cs
	}
}

func WithMaxStreamDuration(d time.Duration) GRPCListenerOpt {
	return func(s *gtcv1alpha1.GRPCListener) {
		s.Spec.MaxStreamDuration = &metav1.Duration{Duration: d}
	}
}

func WithDefaultCluster(l gtcv1alpha1.DefaultCluster) GRPCListenerOpt {
	return func(s *gtcv1alpha1.GRPCListener) {
		s.Spec.DefaultCluster = &l
	}
}

func BuildGRPCListener(name, namespace string, opts ...GRPCListenerOpt) gtcv1alpha1.GRPCListener {
	s := gtcv1alpha1.GRPCListener{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	for _, o := range opts {
		o(&s)
	}

	return s
}
