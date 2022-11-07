package testruntime

import (
	"net"
	"strconv"

	kxdsv1alpha1 "github.com/jlevesy/kxds/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

func WithK8sService(s kxdsv1alpha1.K8sService) LocalityOption {
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

type ClusterOption func(c *kxdsv1alpha1.Cluster)

func WithLocalities(ls ...kxdsv1alpha1.Locality) ClusterOption {
	return func(c *kxdsv1alpha1.Cluster) {
		c.Localities = ls
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

type RouteOption func(r *kxdsv1alpha1.Route)

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
		p.CaseSensitive = v
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

func BuildXDSService(name, namespace, listener string, opts ...XDSServiceOpt) kxdsv1alpha1.XDSService {
	s := kxdsv1alpha1.XDSService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: kxdsv1alpha1.XDSServiceSpec{
			Listener: listener,
		},
	}

	for _, o := range opts {
		o(&s)
	}

	return s
}

func BuildEndpoints(name, namespace string, backends []Backend) corev1.Endpoints {
	subsets := make([]corev1.EndpointSubset, len(backends))

	for i, b := range backends {
		_, p, _ := net.SplitHostPort(b.Listener.Addr().String())
		pp, _ := strconv.Atoi(p)

		subsets[i] = corev1.EndpointSubset{
			Addresses: []corev1.EndpointAddress{
				{
					IP: "127.0.0.1",
				},
			},
			Ports: []corev1.EndpointPort{
				{
					Port: int32(pp),
					Name: "grpc",
				},
			},
		}
	}

	return corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Subsets: subsets,
	}
}
