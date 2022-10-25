package kxds_test

import (
	"context"
	"testing"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/stretchr/testify/require"
	_ "google.golang.org/grpc/xds"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kxdsv1alpha1 "github.com/jlevesy/kxds/api/v1alpha1"
	"github.com/jlevesy/kxds/kxds"
	"github.com/jlevesy/kxds/pkg/testruntime"
)

var (
	grpcPort = kxdsv1alpha1.K8sPort{
		Name: "grpc",
	}
)

func TestReconciller(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	backends, err := testruntime.StartBackends(
		testruntime.Config{
			BackendCount: 10,
		},
	)
	require.NoError(t, err)
	defer backends.Stop()

	var (
		xdsCache = cache.NewSnapshotCache(
			false,
			kxds.DefaultHash,
			nil,
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

		doAssert func(t *testing.T)
	}{
		{
			desc: "single call port by name",
			endpoints: []corev1.Endpoints{
				testruntime.BuildEndpoints("test-service", "default", backends[0:1]),
			},
			xdsServices: []kxdsv1alpha1.XDSService{
				testruntime.BuildXDSService(
					"test-xds",
					"default",
					"echo_server",
					testruntime.WithRoutes(
						testruntime.BuildSingleRoute("default"),
					),
					testruntime.WithClusters(
						testruntime.BuildCluster(
							"default",
							testruntime.WithLocalities(
								testruntime.BuildLocality(
									testruntime.WithK8sService(
										kxdsv1alpha1.K8sService{
											Name: "test-service",
											Port: grpcPort,
										},
									),
								),
							),
						),
					),
				),
			},
			doAssert: testruntime.CallOnce(
				"xds:///echo_server",
				testruntime.NoCallErrors,
				testruntime.AggregateByBackendID(
					testruntime.BackendCalledExact("backend-0", 1),
				),
			),
		},
		{
			desc: "single call port by number",
			endpoints: []corev1.Endpoints{
				testruntime.BuildEndpoints("test-service", "default", backends[0:1]),
			},
			xdsServices: []kxdsv1alpha1.XDSService{
				testruntime.BuildXDSService(
					"test-xds",
					"default",
					"echo_server",
					testruntime.WithRoutes(
						testruntime.BuildSingleRoute("default"),
					),
					testruntime.WithClusters(
						testruntime.BuildCluster(
							"default",
							testruntime.WithLocalities(
								testruntime.BuildLocality(
									testruntime.WithK8sService(
										kxdsv1alpha1.K8sService{
											Name: "test-service",
											Port: kxdsv1alpha1.K8sPort{
												Number: backends[0].PortNumber(),
											},
										},
									),
								),
							),
						),
					),
				),
			},
			doAssert: testruntime.CallOnce(
				"xds:///echo_server",
				testruntime.NoCallErrors,
				testruntime.AggregateByBackendID(
					testruntime.BackendCalledExact("backend-0", 1),
				),
			),
		},
		{
			desc: "locality based wrr",
			endpoints: []corev1.Endpoints{
				testruntime.BuildEndpoints("test-service", "default", backends[0:2]),
				testruntime.BuildEndpoints("test-service-v2", "default", backends[2:4]),
			},
			xdsServices: []kxdsv1alpha1.XDSService{
				testruntime.BuildXDSService(
					"test-xds",
					"default",
					"echo_server",
					testruntime.WithRoutes(
						testruntime.BuildSingleRoute("default"),
					),
					testruntime.WithClusters(
						testruntime.BuildCluster(
							"default",
							testruntime.WithLocalities(
								testruntime.BuildLocality(
									testruntime.WithLocalityWeight(80),
									testruntime.WithK8sService(
										kxdsv1alpha1.K8sService{
											Name: "test-service",
											Port: grpcPort,
										},
									),
								),
								testruntime.BuildLocality(
									testruntime.WithLocalityWeight(20),
									testruntime.WithK8sService(
										kxdsv1alpha1.K8sService{
											Name: "test-service-v2",
											Port: grpcPort,
										},
									),
								),
							),
						),
					),
				),
			},
			doAssert: testruntime.CallN(
				"xds:///echo_server",
				10000,
				testruntime.NoCallErrors,
				testruntime.AggregateByBackendID(
					// 80% of calls
					testruntime.BackendCalledDelta("backend-0", 4000, 500.0),
					testruntime.BackendCalledDelta("backend-1", 4000, 500.0),
					// 20% of calls
					testruntime.BackendCalledDelta("backend-2", 1000, 500.0),
					testruntime.BackendCalledDelta("backend-3", 1000, 500.0),
				),
			),
		},
		{
			desc: "priority fallback",
			endpoints: []corev1.Endpoints{
				// No backends for the test-service in that case.
				testruntime.BuildEndpoints("test-service", "default", backends[0:0]),
				testruntime.BuildEndpoints("test-service-v2", "default", backends[1:2]),
			},
			xdsServices: []kxdsv1alpha1.XDSService{
				testruntime.BuildXDSService(
					"test-xds",
					"default",
					"echo_server",
					testruntime.WithRoutes(
						testruntime.BuildSingleRoute("default"),
					),
					testruntime.WithClusters(
						testruntime.BuildCluster(
							"default",
							testruntime.WithLocalities(
								testruntime.BuildLocality(
									testruntime.WithLocalityPriority(0),
									testruntime.WithK8sService(
										kxdsv1alpha1.K8sService{
											Name: "test-service",
											Port: grpcPort,
										},
									),
								),
								testruntime.BuildLocality(
									testruntime.WithLocalityPriority(1),
									testruntime.WithK8sService(
										kxdsv1alpha1.K8sService{
											Name: "test-service-v2",
											Port: grpcPort,
										},
									),
								),
							),
						),
					),
				),
			},
			doAssert: testruntime.CallOnce(
				"xds:///echo_server",
				testruntime.NoCallErrors,
				testruntime.AggregateByBackendID(
					// No calls for the first set of backends
					testruntime.BackendCalledExact("backend-0", 0),
					// One call for the second backend.
					testruntime.BackendCalledExact("backend-1", 1),
				),
			),
		},
	} {
		t.Run(testCase.desc, func(t *testing.T) {
			var (
				cl = fake.NewClientBuilder().WithLists(
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

			testCase.doAssert(t)
		})
	}
}
