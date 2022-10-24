package kxds_test

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/xds"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	kxdsv1alpha1 "github.com/jlevesy/kxds/api/v1alpha1"
	"github.com/jlevesy/kxds/kxds"
	echo "github.com/jlevesy/kxds/pkg/echoserver/proto"
	"github.com/jlevesy/kxds/pkg/testruntime"
)

func TestReconciller(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runtime, err := testruntime.New(
		testruntime.Config{
			BackendCount: 10,
		},
	)
	require.NoError(t, err)
	defer runtime.Stop()

	var (
		xdsCache = cache.NewSnapshotCache(
			false,
			kxds.DefaultHash,
			nil, // TODO
		)

		server = kxds.NewXDSServer(
			xdsCache,
			kxds.XDSServerConfig{BindAddr: ":18000"},
		)
	)

	go func() {
		err := server.Start(ctx)
		require.NoError(t, err)
	}()

	for _, testCase := range []struct {
		desc        string
		endpoints   []corev1.Endpoints
		xdsServices []kxdsv1alpha1.XDSService

		assertCall func(t *testing.T)
	}{
		{
			desc: "single call",
			endpoints: []corev1.Endpoints{
				genEndpoints("test-service", "default", runtime.Backends),
			},
			xdsServices: []kxdsv1alpha1.XDSService{
				genXDSService("test-xds", "default", "echo_server", "test-service"),
			},
			assertCall: successful("xds:///echo_server"),
		},
	} {
		t.Run(testCase.desc, func(t *testing.T) {
			var (
				cl = runtime.Client.WithLists(
					&kxdsv1alpha1.XDSServiceList{Items: testCase.xdsServices},
					&corev1.EndpointsList{Items: testCase.endpoints},
				).Build()

				cacheReconciller = kxds.NewReconciler(
					cl,
					kxds.NewCacheRefresher(
						xdsCache,
						kxds.DefautHashKey,
					),
				)
			)

			// Flush snapshot state from previous iteration.
			xdsCache.ClearSnapshot(kxds.DefautHashKey)

			_, err := cacheReconciller.Reconcile(
				ctx,
				ctrl.Request{},
			)
			require.NoError(t, err)

			testCase.assertCall(t)
		})
	}
}

func successful(addr string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := grpc.Dial(
			addr,
			grpc.WithTransportCredentials(
				insecure.NewCredentials(),
			),
		)
		require.NoError(t, err)

		defer conn.Close()

		resp, err := echo.NewEchoClient(conn).Echo(context.Background(), &echo.EchoRequest{Payload: "Hello there"})
		require.NoError(t, err)

		t.Log("Received a response from", resp.ServerId)
	}
}

func genEndpoints(name, namespace string, backends []testruntime.Backend) corev1.Endpoints {
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

func genXDSService(name, namespace, listener, serviceName string) kxdsv1alpha1.XDSService {
	return kxdsv1alpha1.XDSService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: kxdsv1alpha1.ServiceSpec{
			Listener: listener,
			Destination: kxdsv1alpha1.Destination{
				Name: serviceName,
			},
		},
	}
}
