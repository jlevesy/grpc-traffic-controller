package testruntime

import (
	"time"

	kxdsv1alpha1 "github.com/jlevesy/kxds/api/kxds/v1alpha1"
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

type ClusterOption func(c *kxdsv1alpha1.Cluster)

func WithMaxRequests(req uint32) ClusterOption {
	return func(c *kxdsv1alpha1.Cluster) {
		c.MaxRequests = &req
	}
}

func WithServiceRef(s kxdsv1alpha1.ServiceRef) ClusterOption {
	return func(c *kxdsv1alpha1.Cluster) {
		c.Service = &s
	}
}

func WithLocalities(l ...kxdsv1alpha1.Locality) ClusterOption {
	return func(c *kxdsv1alpha1.Cluster) {
		c.Localities = l
	}
}

func BuildCluster(name string, opts ...ClusterOption) kxdsv1alpha1.Cluster {
	c := kxdsv1alpha1.Cluster{
		Name: name,
	}

	for _, opt := range opts {
		opt(&c)
	}

	return c
}

type LocalityOption func(l *kxdsv1alpha1.Locality)

func WithLocalityWeight(weight uint32) LocalityOption {
	return func(l *kxdsv1alpha1.Locality) {
		l.Weight = weight
	}
}

func WithLocalityPriority(priority uint32) LocalityOption {
	return func(l *kxdsv1alpha1.Locality) {
		l.Priority = priority
	}
}

func WithLocalityServiceRef(s kxdsv1alpha1.ServiceRef) LocalityOption {
	return func(l *kxdsv1alpha1.Locality) {
		l.Service = &s
	}
}

func BuildLocality(opts ...LocalityOption) kxdsv1alpha1.Locality {
	l := kxdsv1alpha1.Locality{
		Weight: 1,
	}

	for _, opt := range opts {
		opt(&l)
	}

	return l
}

type DefaultClusterOption func(c *kxdsv1alpha1.DefaultCluster)

func BuildDefaultCluster(opts ...DefaultClusterOption) kxdsv1alpha1.DefaultCluster {
	var c kxdsv1alpha1.DefaultCluster

	for _, o := range opts {
		o(&c)
	}

	return c
}

func WithDefaultServiceRef(s kxdsv1alpha1.ServiceRef) DefaultClusterOption {
	return func(c *kxdsv1alpha1.DefaultCluster) {
		c.Service = &s
	}
}

func HeaderInvertMatch(in kxdsv1alpha1.HeaderMatcher) kxdsv1alpha1.HeaderMatcher {
	in.Invert = true
	return in
}

func HeaderExactMatch(name, value string) kxdsv1alpha1.HeaderMatcher {
	return kxdsv1alpha1.HeaderMatcher{
		Name:  name,
		Exact: &value,
	}
}

func HeaderPresentMatch(name string, present bool) kxdsv1alpha1.HeaderMatcher {
	return kxdsv1alpha1.HeaderMatcher{
		Name:    name,
		Present: &present,
	}
}

func HeaderPrefixMatch(name, prefix string) kxdsv1alpha1.HeaderMatcher {
	return kxdsv1alpha1.HeaderMatcher{
		Name:   name,
		Prefix: &prefix,
	}
}

func HeaderSuffixMatch(name, suffix string) kxdsv1alpha1.HeaderMatcher {
	return kxdsv1alpha1.HeaderMatcher{
		Name:   name,
		Suffix: &suffix,
	}
}

type RouteOption func(r *kxdsv1alpha1.Route)

func WithHeaderMatchers(matchers ...kxdsv1alpha1.HeaderMatcher) RouteOption {
	return func(r *kxdsv1alpha1.Route) {
		r.Headers = matchers
	}
}

func WithRouteMaxStreamDuration(d time.Duration) RouteOption {
	return func(r *kxdsv1alpha1.Route) {
		r.MaxStreamDuration = &metav1.Duration{Duration: d}
	}
}

func WithRuntimeFraction(fr kxdsv1alpha1.Fraction) RouteOption {
	return func(r *kxdsv1alpha1.Route) {
		r.RuntimeFraction = &fr
	}
}

func WithClusterRefs(refs ...kxdsv1alpha1.ClusterRef) RouteOption {
	return func(r *kxdsv1alpha1.Route) {
		r.Clusters = refs
	}
}

func WithPathMatcher(pm kxdsv1alpha1.PathMatcher) RouteOption {
	return func(r *kxdsv1alpha1.Route) {
		r.Path = pm
	}
}

func WithCaseSensitive(v bool) RouteOption {
	return func(r *kxdsv1alpha1.Route) {
		r.CaseSensitive = v
	}
}

func BuildRoute(opts ...RouteOption) kxdsv1alpha1.Route {
	r := kxdsv1alpha1.Route{
		Path: kxdsv1alpha1.PathMatcher{
			Prefix: "/",
		},
	}

	for _, opt := range opts {
		opt(&r)
	}

	return r
}

func BuildSingleRoute(clusterName string) kxdsv1alpha1.Route {
	return BuildRoute(
		WithClusterRefs(
			kxdsv1alpha1.ClusterRef{
				Name:   clusterName,
				Weight: 1,
			},
		),
	)
}

type XDSServiceOpt func(s *kxdsv1alpha1.XDSService)

func WithFilters(fs ...kxdsv1alpha1.Filter) XDSServiceOpt {
	return func(s *kxdsv1alpha1.XDSService) {
		s.Spec.Filters = fs
	}
}

func WithRoutes(rs ...kxdsv1alpha1.Route) XDSServiceOpt {
	return func(s *kxdsv1alpha1.XDSService) {
		s.Spec.Routes = rs
	}
}

func WithClusters(cs ...kxdsv1alpha1.Cluster) XDSServiceOpt {
	return func(s *kxdsv1alpha1.XDSService) {
		s.Spec.Clusters = cs
	}
}

func WithMaxStreamDuration(d time.Duration) XDSServiceOpt {
	return func(s *kxdsv1alpha1.XDSService) {
		s.Spec.MaxStreamDuration = &metav1.Duration{Duration: d}
	}
}

func WithDefaultCluster(l kxdsv1alpha1.DefaultCluster) XDSServiceOpt {
	return func(s *kxdsv1alpha1.XDSService) {
		s.Spec.DefaultCluster = &l
	}
}

func BuildXDSService(name, namespace string, opts ...XDSServiceOpt) kxdsv1alpha1.XDSService {
	s := kxdsv1alpha1.XDSService{
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
