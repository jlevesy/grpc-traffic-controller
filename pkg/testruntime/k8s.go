package testruntime

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"testing"
	"time"

	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeinformers "k8s.io/client-go/informers"
	kubefake "k8s.io/client-go/kubernetes/fake"

	gtcv1alpha1 "github.com/jlevesy/grpc-traffic-controller/api/gtc/v1alpha1"
	gtcfake "github.com/jlevesy/grpc-traffic-controller/client/clientset/versioned/fake"
	gtcinformers "github.com/jlevesy/grpc-traffic-controller/client/informers/externalversions"
	"github.com/stretchr/testify/require"
)

type FakeK8s struct {
	K8s          *kubefake.Clientset
	K8sInformers kubeinformers.SharedInformerFactory

	GTCApi       *gtcfake.Clientset
	GTCInformers gtcinformers.SharedInformerFactory
}

func NewFakeK8s(t *testing.T, listeners []gtcv1alpha1.GRPCListener, endpointSlices []discoveryv1.EndpointSlice) FakeK8s {
	t.Helper()

	var (
		gtcClientSet = gtcfake.NewSimpleClientset(grpcListenersToRuntimeObjects(listeners)...)
		gtcInformers = gtcinformers.NewSharedInformerFactory(
			gtcClientSet,
			60*time.Second,
		)
		k8sClientSet = kubefake.NewSimpleClientset(endpointSlicesToRuntimeObjects(endpointSlices)...)
		k8sInformers = kubeinformers.NewSharedInformerFactory(
			k8sClientSet,
			60*time.Second,
		)
	)

	return FakeK8s{
		K8s:          k8sClientSet,
		K8sInformers: k8sInformers,
		GTCApi:       gtcClientSet,
		GTCInformers: gtcInformers,
	}
}

func (f *FakeK8s) Start(ctx context.Context, t *testing.T) {
	t.Helper()

	f.K8sInformers.Start(ctx.Done())
	f.GTCInformers.Start(ctx.Done())

	err := checkInformerSync(f.K8sInformers.WaitForCacheSync(ctx.Done()))
	require.NoError(t, err)

	err = checkInformerSync(f.GTCInformers.WaitForCacheSync(ctx.Done()))
	require.NoError(t, err)
}

func checkInformerSync(syncResult map[reflect.Type]bool) error {
	if len(syncResult) == 0 {
		return errors.New("empty sync result")
	}

	for typ, ok := range syncResult {
		if !ok {
			return fmt.Errorf("Cache sync failed for %s, exiting", typ.String())
		}
	}

	return nil
}

func AppendEndpointSlices(ss ...[]discoveryv1.EndpointSlice) []discoveryv1.EndpointSlice {
	var r []discoveryv1.EndpointSlice

	for _, s := range ss {
		r = append(r, s...)
	}

	return r
}

func BuildEndpointSlices(name, namespace string, backends []Backend) []discoveryv1.EndpointSlice {
	slices := make([]discoveryv1.EndpointSlice, len(backends))

	for i, b := range backends {
		slices[i] = buildEndpointSlice(i, name, namespace, b)
	}

	return slices
}

func buildEndpointSlice(id int, name, namespace string, backend Backend) discoveryv1.EndpointSlice {
	_, p, _ := net.SplitHostPort(backend.Listener.Addr().String())
	pp, _ := strconv.Atoi(p)

	return discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", name, id),
			Namespace: namespace,
			Labels: map[string]string{
				"kubernetes.io/service-name": name,
			},
		},
		Ports: []discoveryv1.EndpointPort{
			{
				Port: Ptr(int32(pp)),
				Name: Ptr("grpc"),
			},
		},
		Endpoints: []discoveryv1.Endpoint{
			{
				Addresses: []string{"127.0.0.1"},
				Conditions: discoveryv1.EndpointConditions{
					Ready: Ptr(true),
				},
			},
		},
	}
}

func grpcListenersToRuntimeObjects(listeners []gtcv1alpha1.GRPCListener) []runtime.Object {
	res := make([]runtime.Object, len(listeners))

	for i, s := range listeners {
		s := s
		res[i] = &s
	}

	return res
}

func endpointSlicesToRuntimeObjects(endpointSlices []discoveryv1.EndpointSlice) []runtime.Object {
	res := make([]runtime.Object, len(endpointSlices))

	for i, ep := range endpointSlices {
		ep := ep
		res[i] = &ep
	}

	return res
}
