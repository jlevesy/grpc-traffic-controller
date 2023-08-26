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

type BackendOption func(c *gtcv1alpha1.Backend)

func WithMaxRequests(req uint32) BackendOption {
	return func(c *gtcv1alpha1.Backend) {
		c.MaxRequests = &req
	}
}

func WithServiceRef(s gtcv1alpha1.ServiceRef) BackendOption {
	return func(c *gtcv1alpha1.Backend) {
		c.Service = &s
	}
}

func WithLocalities(l ...gtcv1alpha1.Locality) BackendOption {
	return func(c *gtcv1alpha1.Backend) {
		c.Localities = l
	}
}

func WithBackendInterceptorOverrides(is ...gtcv1alpha1.Interceptor) BackendOption {
	return func(c *gtcv1alpha1.Backend) {
		c.Interceptors = is
	}
}

func BuildBackend(opts ...BackendOption) gtcv1alpha1.Backend {
	c := gtcv1alpha1.Backend{Weight: 1}

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

func MetadataInvertMatch(in gtcv1alpha1.MetadataMatcher) gtcv1alpha1.MetadataMatcher {
	in.Invert = true
	return in
}

func MetadataExactMatch(name, value string) gtcv1alpha1.MetadataMatcher {
	return gtcv1alpha1.MetadataMatcher{
		Name:  name,
		Exact: &value,
	}
}

func MetadataPresentMatch(name string, present bool) gtcv1alpha1.MetadataMatcher {
	return gtcv1alpha1.MetadataMatcher{
		Name:    name,
		Present: &present,
	}
}

func MetadataPrefixMatch(name, prefix string) gtcv1alpha1.MetadataMatcher {
	return gtcv1alpha1.MetadataMatcher{
		Name:   name,
		Prefix: &prefix,
	}
}

func MetadataSuffixMatch(name, suffix string) gtcv1alpha1.MetadataMatcher {
	return gtcv1alpha1.MetadataMatcher{
		Name:   name,
		Suffix: &suffix,
	}
}

type RouteMatcherOption func(m *gtcv1alpha1.RouteMatcher)

func WithMethodMatcher(namespace, service, method string) RouteMatcherOption {
	return func(m *gtcv1alpha1.RouteMatcher) {
		m.Method = &gtcv1alpha1.MethodMatcher{
			Namespace: namespace,
			Service:   service,
			Method:    method,
		}
	}
}

func WithFractionMatcher(fr gtcv1alpha1.Fraction) RouteMatcherOption {
	return func(r *gtcv1alpha1.RouteMatcher) {
		r.Fraction = &fr
	}
}

func WithServiceMatcher(namespace, service string) RouteMatcherOption {
	return func(m *gtcv1alpha1.RouteMatcher) {
		m.Service = &gtcv1alpha1.ServiceMatcher{
			Namespace: namespace,
			Service:   service,
		}
	}
}

func WithNamespaceMatcher(namespace string) RouteMatcherOption {
	return func(m *gtcv1alpha1.RouteMatcher) {
		m.Namespace = &namespace
	}
}

func WithMetadataMatchers(mms ...gtcv1alpha1.MetadataMatcher) RouteMatcherOption {
	return func(m *gtcv1alpha1.RouteMatcher) {
		m.Metadata = mms
	}
}

func BuildRouteMatcher(opts ...RouteMatcherOption) gtcv1alpha1.RouteMatcher {
	var m gtcv1alpha1.RouteMatcher

	for _, o := range opts {
		o(&m)
	}

	return m
}

type RouteOption func(r *gtcv1alpha1.Route)

func WithRouteMatcher(m gtcv1alpha1.RouteMatcher) RouteOption {
	return func(r *gtcv1alpha1.Route) {
		r.Matcher = &m
	}
}

func WithRouteMaxStreamDuration(d time.Duration) RouteOption {
	return func(r *gtcv1alpha1.Route) {
		r.MaxStreamDuration = &metav1.Duration{Duration: d}
	}
}

func WithBackends(backends ...gtcv1alpha1.Backend) RouteOption {
	return func(r *gtcv1alpha1.Route) {
		r.Backends = backends
	}
}

func WithRouteInterceptorOverrides(overrides ...gtcv1alpha1.Interceptor) RouteOption {
	return func(r *gtcv1alpha1.Route) {
		r.Interceptors = overrides
	}
}

func BuildRoute(opts ...RouteOption) gtcv1alpha1.Route {
	r := gtcv1alpha1.Route{}

	for _, opt := range opts {
		opt(&r)
	}

	return r
}

type GRPCListenerOpt func(s *gtcv1alpha1.GRPCListener)

func WithInterceptors(fs ...gtcv1alpha1.Interceptor) GRPCListenerOpt {
	return func(s *gtcv1alpha1.GRPCListener) {
		s.Spec.Interceptors = fs
	}
}

func WithRoutes(rs ...gtcv1alpha1.Route) GRPCListenerOpt {
	return func(s *gtcv1alpha1.GRPCListener) {
		s.Spec.Routes = rs
	}
}

func WithMaxStreamDuration(d time.Duration) GRPCListenerOpt {
	return func(s *gtcv1alpha1.GRPCListener) {
		s.Spec.MaxStreamDuration = &metav1.Duration{Duration: d}
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
