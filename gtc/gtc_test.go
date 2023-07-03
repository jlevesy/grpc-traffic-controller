package gtc_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	_ "google.golang.org/grpc/xds"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gtcv1alpha1 "github.com/jlevesy/grpc-traffic-controller/api/gtc/v1alpha1"
	"github.com/jlevesy/grpc-traffic-controller/gtc"
	"github.com/jlevesy/grpc-traffic-controller/pkg/testruntime"
)

const (
	defaultNamespace = "default"
	serviceNameV1    = "test-service-v1"
	serviceNameV2    = "test-service-v2"
)

var (
	grpcPort = gtcv1alpha1.PortRef{
		Name: "grpc",
	}

	v1v2ClusterTopology = testruntime.WithClusters(
		testruntime.BuildCluster(
			"v2",
			testruntime.WithServiceRef(
				gtcv1alpha1.ServiceRef{
					Name: serviceNameV2,
					Port: grpcPort,
				},
			),
		),
		testruntime.BuildCluster(
			"v1",
			testruntime.WithServiceRef(
				gtcv1alpha1.ServiceRef{
					Name: serviceNameV1,
					Port: grpcPort,
				},
			),
		),
	)
)

func TestServer(t *testing.T) {
	for _, testCase := range []struct {
		desc                string
		backendCount        int
		buildEndpointSlices func(backends []testruntime.Backend) []discoveryv1.EndpointSlice
		buildXDSServices    func(backends []testruntime.Backend) []gtcv1alpha1.XDSService
		buildCallContext    func(t *testing.T) *testruntime.CallContext
		setBackendsBehavior func(t *testing.T, bs testruntime.Backends)
		doAssertPreUpdate   func(t *testing.T, callCtx *testruntime.CallContext)
		updateResources     func(t *testing.T, k8s testruntime.FakeK8s, backends []testruntime.Backend)
		doAssertPostUpdate  func(t *testing.T, callCtx *testruntime.CallContext)
	}{
		{
			desc:         "single call port by name",
			backendCount: 1,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.BuildEndpointSlices(
					serviceNameV1,
					defaultNamespace,
					backends[0:1],
				)
			},
			buildXDSServices: func([]testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildSingleRoute("default"),
						),
						testruntime.WithClusters(
							testruntime.BuildCluster(
								"default",
								testruntime.WithServiceRef(
									gtcv1alpha1.ServiceRef{
										Name: serviceNameV1,
										Port: grpcPort,
									},
								),
							),
						),
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.CallOnce(
				testruntime.BuildCaller(
					testruntime.MethodEcho,
				),
				testruntime.NoCallErrors,
				testruntime.CountByBackendID(
					testruntime.AssertCount("backend-0", 1),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "single call port by number",
			backendCount: 1,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.BuildEndpointSlices(
					serviceNameV1,
					defaultNamespace,
					backends[0:1],
				)
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildSingleRoute("default"),
						),
						testruntime.WithClusters(
							testruntime.BuildCluster(
								"default",
								testruntime.WithServiceRef(
									gtcv1alpha1.ServiceRef{
										Name: serviceNameV1,
										Port: gtcv1alpha1.PortRef{
											Number: backends[0].PortNumber(),
										},
									},
								),
							),
						),
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.CallOnce(
				testruntime.BuildCaller(
					testruntime.MethodEcho,
				),
				testruntime.NoCallErrors,
				testruntime.CountByBackendID(
					testruntime.AssertCount("backend-0", 1),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "single call default cluster",
			backendCount: 1,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.BuildEndpointSlices(
					serviceNameV1,
					defaultNamespace,
					backends[0:1],
				)
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithDefaultCluster(
							testruntime.BuildDefaultCluster(
								testruntime.WithDefaultServiceRef(
									gtcv1alpha1.ServiceRef{
										Name: serviceNameV1,
										Port: grpcPort,
									},
								),
							),
						),
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.CallOnce(
				testruntime.BuildCaller(
					testruntime.MethodEcho,
				),
				testruntime.NoCallErrors,
				testruntime.CountByBackendID(
					testruntime.AssertCount("backend-0", 1),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "cross namespace",
			backendCount: 1,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.BuildEndpointSlices(
					serviceNameV1,
					"some-app",
					backends[0:1],
				)
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildSingleRoute("default"),
						),
						testruntime.WithClusters(
							testruntime.BuildCluster(
								"default",
								testruntime.WithServiceRef(
									gtcv1alpha1.ServiceRef{
										Name:      serviceNameV1,
										Namespace: "some-app",
										Port:      grpcPort,
									},
								),
							),
						),
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.CallOnce(
				testruntime.BuildCaller(
					testruntime.MethodEcho,
				),
				testruntime.NoCallErrors,
				testruntime.CountByBackendID(
					testruntime.AssertCount("backend-0", 1),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "locality based wrr",
			backendCount: 4,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.AppendEndpointSlices(
					testruntime.BuildEndpointSlices("test-service", "default", backends[0:2]),
					testruntime.BuildEndpointSlices("test-service-v2", "default", backends[2:4]),
				)
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildSingleRoute("default"),
						),
						testruntime.WithClusters(
							testruntime.BuildCluster(
								"default",
								testruntime.WithLocalities(
									testruntime.BuildLocality(
										testruntime.WithLocalityWeight(80),
										testruntime.WithLocalityServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: "test-service",
												Port: grpcPort,
											},
										),
									),
									testruntime.BuildLocality(
										testruntime.WithLocalityWeight(20),
										testruntime.WithLocalityServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: "test-service-v2",
												Port: grpcPort,
											},
										),
									),
								),
							),
						),
					),
				}
			},
			setBackendsBehavior: answer,
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			doAssertPreUpdate: testruntime.CallN(
				testruntime.BuildCaller(
					testruntime.MethodEcho,
				),
				10000,
				testruntime.NoCallErrors,
				testruntime.CountByBackendID(
					// 80% of calls
					testruntime.AssertCountWithinDelta("backend-0", 4000, 500.0),
					testruntime.AssertCountWithinDelta("backend-1", 4000, 500.0),
					// 20% of calls
					testruntime.AssertCountWithinDelta("backend-2", 1000, 500.0),
					testruntime.AssertCountWithinDelta("backend-3", 1000, 500.0),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "priority fallback",
			backendCount: 2,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.AppendEndpointSlices(
					// No backend for this service.
					testruntime.BuildEndpointSlices("test-service", "default", backends[0:0]),
					testruntime.BuildEndpointSlices("test-service-v2", "default", backends[1:2]),
				)
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildSingleRoute("default"),
						),
						testruntime.WithClusters(
							testruntime.BuildCluster(
								"default",
								testruntime.WithLocalities(
									testruntime.BuildLocality(
										testruntime.WithLocalityPriority(0),
										testruntime.WithLocalityServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: "test-service",
												Port: grpcPort,
											},
										),
									),
									testruntime.BuildLocality(
										testruntime.WithLocalityPriority(1),
										testruntime.WithLocalityServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: "test-service-v2",
												Port: grpcPort,
											},
										),
									),
								),
							),
						),
					),
				}
			},
			setBackendsBehavior: answer,
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			doAssertPreUpdate: testruntime.CallOnce(
				testruntime.BuildCaller(
					testruntime.MethodEcho,
				),
				testruntime.NoCallErrors,
				testruntime.CountByBackendID(
					// No calls for the first set of backends
					testruntime.AssertCount("backend-0", 0),
					// One call for the second backend.
					testruntime.AssertCount("backend-1", 1),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "exact path matching",
			backendCount: 2,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.AppendEndpointSlices(
					testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					testruntime.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildRoute(
								testruntime.WithPathMatcher(
									gtcv1alpha1.PathMatcher{
										Path: "/echo.Echo/EchoPremium",
									},
								),
								testruntime.WithClusterRefs(
									gtcv1alpha1.ClusterRef{
										Name:   "v2",
										Weight: 1,
									},
								),
							),
							testruntime.BuildSingleRoute("v1"),
						),
						v1v2ClusterTopology,
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.MultiAssert(
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEchoPremium,
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						// One call for the second backend, because we're calling premium.
						testruntime.AssertCount("backend-1", 1),
						testruntime.AssertCount("backend-0", 0),
					),
				),
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						// No calls for the first set of backends
						// First backend should get a call.
						testruntime.AssertCount("backend-0", 1),
						testruntime.AssertCount("backend-1", 0),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "prefix path matching",
			backendCount: 2,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.AppendEndpointSlices(
					testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					testruntime.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildRoute(
								testruntime.WithPathMatcher(
									gtcv1alpha1.PathMatcher{
										Prefix: "/echo.Echo/EchoP",
									},
								),
								testruntime.WithClusterRefs(
									gtcv1alpha1.ClusterRef{
										Name:   "v2",
										Weight: 1,
									},
								),
							),
							testruntime.BuildSingleRoute("v1"),
						),
						v1v2ClusterTopology,
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.MultiAssert(
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEchoPremium,
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						// One call for the second backend, because we're calling premium.
						testruntime.AssertCount("backend-1", 1),
						testruntime.AssertCount("backend-0", 0),
					),
				),
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						// No calls for the first set of backends
						// First backend should get a call.
						testruntime.AssertCount("backend-0", 1),
						testruntime.AssertCount("backend-1", 0),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "regexp path matching",
			backendCount: 2,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.AppendEndpointSlices(
					testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					testruntime.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildRoute(
								testruntime.WithPathMatcher(
									gtcv1alpha1.PathMatcher{
										Prefix: "/echo.Echo/EchoP",
										Regex: &gtcv1alpha1.RegexMatcher{
											Regex:  ".*/EchoPremium",
											Engine: "re2",
										},
									},
								),
								testruntime.WithClusterRefs(
									gtcv1alpha1.ClusterRef{
										Name:   "v2",
										Weight: 1,
									},
								),
							),
							testruntime.BuildSingleRoute("v1"),
						),
						v1v2ClusterTopology,
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.MultiAssert(
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEchoPremium,
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						// One call for the second backend, because we're calling premium.
						testruntime.AssertCount("backend-1", 1),
						testruntime.AssertCount("backend-0", 0),
					),
				),
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						// No calls for the first set of backends
						// First backend should get a call.
						testruntime.AssertCount("backend-0", 1),
						testruntime.AssertCount("backend-1", 0),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "case insensitive path matching",
			backendCount: 2,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.AppendEndpointSlices(
					testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					testruntime.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildRoute(
								testruntime.WithCaseSensitive(false),
								testruntime.WithPathMatcher(
									gtcv1alpha1.PathMatcher{
										Prefix: "/echo.echo/echop",
									},
								),
								testruntime.WithClusterRefs(
									gtcv1alpha1.ClusterRef{
										Name:   "v2",
										Weight: 1,
									},
								),
							),
							testruntime.BuildSingleRoute("v1"),
						),
						v1v2ClusterTopology,
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.MultiAssert(
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEchoPremium,
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						// One call for the second backend, because we're calling premium.
						testruntime.AssertCount("backend-1", 1),
						testruntime.AssertCount("backend-0", 0),
					),
				),
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						// No calls for the first set of backends
						// First backend should get a call.
						testruntime.AssertCount("backend-0", 1),
						testruntime.AssertCount("backend-1", 0),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "header invert matching",
			backendCount: 2,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.AppendEndpointSlices(
					testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					testruntime.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildRoute(
								testruntime.WithHeaderMatchers(
									testruntime.HeaderInvertMatch(
										testruntime.HeaderExactMatch(
											"x-variant",
											"Awesome",
										),
									),
								),
								testruntime.WithClusterRefs(
									gtcv1alpha1.ClusterRef{
										Name:   "v2",
										Weight: 1,
									},
								),
							),
							testruntime.BuildSingleRoute("v1"),
						),
						v1v2ClusterTopology,
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.MultiAssert(
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
						testruntime.WithMetadata(
							map[string]string{
								"x-variant": "Awesome",
							},
						),
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						testruntime.AssertCount("backend-1", 0),
						testruntime.AssertCount("backend-0", 1),
					),
				),
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
						testruntime.WithMetadata(
							map[string]string{
								"x-variant": "NotAwesome",
							},
						),
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						testruntime.AssertCount("backend-0", 0),
						testruntime.AssertCount("backend-1", 1),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "header exact matching",
			backendCount: 2,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.AppendEndpointSlices(
					testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					testruntime.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildRoute(
								testruntime.WithHeaderMatchers(
									testruntime.HeaderExactMatch(
										"x-variant",
										"Awesome",
									),
								),
								testruntime.WithClusterRefs(
									gtcv1alpha1.ClusterRef{
										Name:   "v2",
										Weight: 1,
									},
								),
							),
							testruntime.BuildSingleRoute("v1"),
						),
						v1v2ClusterTopology,
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.MultiAssert(
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
						testruntime.WithMetadata(
							map[string]string{
								// Gotha, metadata keys are lowercased.
								"x-variant": "Awesome",
							},
						),
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						// One call for the second backend, because we're calling premium.
						testruntime.AssertCount("backend-1", 1),
						testruntime.AssertCount("backend-0", 0),
					),
				),
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						// No calls for the first set of backends
						// First backend should get a call.
						testruntime.AssertCount("backend-0", 1),
						testruntime.AssertCount("backend-1", 0),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "header safe regex match",
			backendCount: 2,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.AppendEndpointSlices(
					testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					testruntime.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildRoute(
								testruntime.WithHeaderMatchers(
									gtcv1alpha1.HeaderMatcher{
										Name: "x-variant",
										Regex: &gtcv1alpha1.RegexMatcher{
											Regex:  "Awe.*",
											Engine: "re2",
										},
									},
								),
								testruntime.WithClusterRefs(
									gtcv1alpha1.ClusterRef{
										Name:   "v2",
										Weight: 1,
									},
								),
							),
							testruntime.BuildSingleRoute("v1"),
						),
						v1v2ClusterTopology,
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.MultiAssert(
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
						testruntime.WithMetadata(
							map[string]string{
								"x-variant": "Awesome",
							},
						),
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						// One call for the second backend, because we're calling premium.
						testruntime.AssertCount("backend-1", 1),
						testruntime.AssertCount("backend-0", 0),
					),
				),
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						// No calls for the first set of backends
						// First backend should get a call.
						testruntime.AssertCount("backend-0", 1),
						testruntime.AssertCount("backend-1", 0),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "header range match",
			backendCount: 2,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.AppendEndpointSlices(
					testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					testruntime.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildRoute(
								testruntime.WithHeaderMatchers(
									gtcv1alpha1.HeaderMatcher{
										Name: "x-variant",
										Range: &gtcv1alpha1.RangeMatcher{
											Start: 10,
											End:   20,
										},
									},
								),
								testruntime.WithClusterRefs(
									gtcv1alpha1.ClusterRef{
										Name:   "v2",
										Weight: 1,
									},
								),
							),
							testruntime.BuildSingleRoute("v1"),
						),
						v1v2ClusterTopology,
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.MultiAssert(
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
						testruntime.WithMetadata(
							map[string]string{
								// In range, call backend 1.
								"x-variant": "12",
							},
						),
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						testruntime.AssertCount("backend-1", 1),
						testruntime.AssertCount("backend-0", 0),
					),
				),
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
						testruntime.WithMetadata(
							map[string]string{
								// Out of bound, call backend-0.
								"x-variant": "9",
							},
						),
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						testruntime.AssertCount("backend-0", 1),
						testruntime.AssertCount("backend-1", 0),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "header present match",
			backendCount: 2,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.AppendEndpointSlices(
					testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					testruntime.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildRoute(
								testruntime.WithHeaderMatchers(
									testruntime.HeaderPresentMatch(
										"x-variant",
										true,
									),
								),
								testruntime.WithClusterRefs(
									gtcv1alpha1.ClusterRef{
										Name:   "v2",
										Weight: 1,
									},
								),
							),
							testruntime.BuildSingleRoute("v1"),
						),
						v1v2ClusterTopology,
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.MultiAssert(
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
						testruntime.WithMetadata(
							map[string]string{
								// Header is present, send to v2.
								"x-variant": "wooop",
							},
						),
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						testruntime.AssertCount("backend-1", 1),
						testruntime.AssertCount("backend-0", 0),
					),
				),
				testruntime.CallOnce(
					testruntime.BuildCaller(testruntime.MethodEcho),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						testruntime.AssertCount("backend-0", 1),
						testruntime.AssertCount("backend-1", 0),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "header prefix match",
			backendCount: 2,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.AppendEndpointSlices(
					testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					testruntime.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildRoute(
								testruntime.WithHeaderMatchers(
									testruntime.HeaderPrefixMatch(
										"x-variant",
										"wo",
									),
								),
								testruntime.WithClusterRefs(
									gtcv1alpha1.ClusterRef{
										Name:   "v2",
										Weight: 1,
									},
								),
							),
							testruntime.BuildSingleRoute("v1"),
						),
						v1v2ClusterTopology,
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.MultiAssert(
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
						testruntime.WithMetadata(
							map[string]string{
								// Header has the prefix wo, send to v2.
								"x-variant": "wooop",
							},
						),
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						testruntime.AssertCount("backend-1", 1),
						testruntime.AssertCount("backend-0", 0),
					),
				),
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
						testruntime.WithMetadata(
							map[string]string{
								// Header has not the prefix wo, send to v1.
								"x-variant": "not",
							},
						),
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						testruntime.AssertCount("backend-0", 1),
						testruntime.AssertCount("backend-1", 0),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "header suffix match",
			backendCount: 2,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.AppendEndpointSlices(
					testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					testruntime.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildRoute(
								testruntime.WithHeaderMatchers(
									testruntime.HeaderSuffixMatch(
										"x-variant",
										"oop",
									),
								),
								testruntime.WithClusterRefs(
									gtcv1alpha1.ClusterRef{
										Name:   "v2",
										Weight: 1,
									},
								),
							),
							testruntime.BuildSingleRoute("v1"),
						),
						v1v2ClusterTopology,
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.MultiAssert(
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
						testruntime.WithMetadata(
							map[string]string{
								// Header has the sufix oop, send to v2.
								"x-variant": "wooop",
							},
						),
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						testruntime.AssertCount("backend-1", 1),
						testruntime.AssertCount("backend-0", 0),
					),
				),
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
						testruntime.WithMetadata(
							map[string]string{
								// Header has not the suffix oop, send to v1.
								"x-variant": "not",
							},
						),
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						testruntime.AssertCount("backend-0", 1),
						testruntime.AssertCount("backend-1", 0),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "runtime fraction traffic splitting",
			backendCount: 2,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.AppendEndpointSlices(
					testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					testruntime.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildRoute(
								// 20.00% of the traffic will go to v2.
								testruntime.WithRuntimeFraction(
									gtcv1alpha1.Fraction{
										Numerator:   20,
										Denominator: "hundred",
									},
								),
								testruntime.WithClusterRefs(
									gtcv1alpha1.ClusterRef{
										Name:   "v2",
										Weight: 1,
									},
								),
							),
							testruntime.BuildSingleRoute("v1"),
						),
						v1v2ClusterTopology,
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.CallN(
				testruntime.BuildCaller(
					testruntime.MethodEcho,
				),
				10000,
				testruntime.NoCallErrors,
				testruntime.CountByBackendID(
					testruntime.AssertCountWithinDelta("backend-1", 2000, 500.0),
					testruntime.AssertCountWithinDelta("backend-0", 8000, 500.0),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "max stream duration",
			backendCount: 1,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1])
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(testruntime.BuildSingleRoute("default")),
						testruntime.WithMaxStreamDuration(50*time.Millisecond),
						testruntime.WithClusters(
							testruntime.BuildCluster(
								"default",
								testruntime.WithServiceRef(
									gtcv1alpha1.ServiceRef{
										Name: serviceNameV1,
										Port: grpcPort,
									},
								),
							),
						),
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: hang(10 * time.Second),
			doAssertPreUpdate: testruntime.WithinDelay(
				time.Second,
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
					),
					testruntime.MustFail,
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "max stream duration on route",
			backendCount: 1,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1])
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildRoute(
								testruntime.WithRouteMaxStreamDuration(50*time.Millisecond),
								testruntime.WithClusterRefs(
									gtcv1alpha1.ClusterRef{
										Name:   "default",
										Weight: 1,
									},
								),
							),
						),
						testruntime.WithClusters(
							testruntime.BuildCluster(
								"default",
								testruntime.WithServiceRef(
									gtcv1alpha1.ServiceRef{
										Name: serviceNameV1,
										Port: grpcPort,
									},
								),
							),
						),
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: hang(10 * time.Second),
			// TODO(jly): this is weird, this test takes 10s when running after max_stream_duration
			// but only 1s  when being run standalone. Something here is fishy.
			doAssertPreUpdate: testruntime.WithinDelay(
				time.Second,
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
					),
					testruntime.MustFail,
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "max requests on cluster",
			backendCount: 1,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1])
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildSingleRoute("default"),
						),
						testruntime.WithClusters(
							testruntime.BuildCluster(
								"default",
								testruntime.WithMaxRequests(1),
								testruntime.WithServiceRef(
									gtcv1alpha1.ServiceRef{
										Name: serviceNameV1,
										Port: grpcPort,
									},
								),
							),
						),
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: hang(1 * time.Second),
			doAssertPreUpdate: testruntime.CallNParallel(
				testruntime.BuildCaller(
					testruntime.MethodEcho,
				),
				10,
				testruntime.AggregateByError(
					testruntime.AssertCount("ok", 1),
					testruntime.AssertAggregatedValuePartial(
						"rpc error: code = Unavailable desc = max requests 1 exceeded",
						9,
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "fixed delay injection",
			backendCount: 1,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1])
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildSingleRoute("default"),
						),
						testruntime.WithFilters(
							gtcv1alpha1.Filter{
								Fault: &gtcv1alpha1.FaultFilter{
									Delay: &gtcv1alpha1.FaultDelay{
										Fixed: testruntime.DurationPtr(500 * time.Millisecond),
										Percentage: &gtcv1alpha1.Fraction{
											Numerator:   100,
											Denominator: "hundred",
										},
									},
								},
							},
						),
						testruntime.WithClusters(
							testruntime.BuildCluster(
								"default",
								testruntime.WithServiceRef(
									gtcv1alpha1.ServiceRef{
										Name: serviceNameV1,
										Port: grpcPort,
									},
								),
							),
						),
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.ExceedDelay(
				200*time.Millisecond,
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
					),
					testruntime.NoCallErrors,
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "header delay injection",
			backendCount: 1,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1])
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildSingleRoute("default"),
						),
						testruntime.WithFilters(
							gtcv1alpha1.Filter{
								Fault: &gtcv1alpha1.FaultFilter{
									Delay: &gtcv1alpha1.FaultDelay{
										Header: &gtcv1alpha1.HeaderFault{},
										Percentage: &gtcv1alpha1.Fraction{
											Numerator:   100,
											Denominator: "hundred",
										},
									},
								},
							},
						),
						testruntime.WithClusters(
							testruntime.BuildCluster(
								"default",
								testruntime.WithServiceRef(
									gtcv1alpha1.ServiceRef{
										Name: serviceNameV1,
										Port: grpcPort,
									},
								),
							),
						),
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.ExceedDelay(
				200*time.Millisecond,
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
						testruntime.WithMetadata(
							map[string]string{
								// Delay by 500ms.
								"x-envoy-fault-delay-request": "500",
							},
						),
					),
					testruntime.NoCallErrors,
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "abort injection http",
			backendCount: 1,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1])
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildSingleRoute("default"),
						),
						testruntime.WithFilters(
							gtcv1alpha1.Filter{
								Fault: &gtcv1alpha1.FaultFilter{
									Abort: &gtcv1alpha1.FaultAbort{
										HTTPStatus: testruntime.Ptr(uint32(404)),
										Percentage: &gtcv1alpha1.Fraction{
											Numerator:   100,
											Denominator: "hundred",
										},
									},
								},
							},
						),
						testruntime.WithClusters(
							testruntime.BuildCluster(
								"default",
								testruntime.WithServiceRef(
									gtcv1alpha1.ServiceRef{
										Name: serviceNameV1,
										Port: grpcPort,
									},
								),
							),
						),
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.CallOnce(
				testruntime.BuildCaller(
					testruntime.MethodEcho,
				),
				testruntime.MustFail,
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "abort injection grpc",
			backendCount: 1,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1])
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildSingleRoute("default"),
						),
						testruntime.WithFilters(
							gtcv1alpha1.Filter{
								Fault: &gtcv1alpha1.FaultFilter{
									Abort: &gtcv1alpha1.FaultAbort{
										GRPCStatus: testruntime.Ptr(uint32(4)),
										Percentage: &gtcv1alpha1.Fraction{
											Numerator:   100,
											Denominator: "hundred",
										},
									},
								},
							},
						),
						testruntime.WithClusters(
							testruntime.BuildCluster(
								"default",
								testruntime.WithServiceRef(
									gtcv1alpha1.ServiceRef{
										Name: serviceNameV1,
										Port: grpcPort,
									},
								),
							),
						),
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.CallOnce(
				testruntime.BuildCaller(
					testruntime.MethodEcho,
				),
				testruntime.MustFail,
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "abort header grpc",
			backendCount: 1,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.BuildEndpointSlices(serviceNameV1, "default", backends[0:1])
			},
			buildXDSServices: func(backends []testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildSingleRoute("default"),
						),
						testruntime.WithFilters(
							gtcv1alpha1.Filter{
								Fault: &gtcv1alpha1.FaultFilter{
									Abort: &gtcv1alpha1.FaultAbort{
										Header: &gtcv1alpha1.HeaderFault{},
										Percentage: &gtcv1alpha1.Fraction{
											Numerator:   100,
											Denominator: "hundred",
										},
									},
								},
							},
						),
						testruntime.WithClusters(
							testruntime.BuildCluster(
								"default",
								testruntime.WithServiceRef(
									gtcv1alpha1.ServiceRef{
										Name: serviceNameV1,
										Port: grpcPort,
									},
								),
							),
						),
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.CallOnce(
				testruntime.BuildCaller(
					testruntime.MethodEcho,
					testruntime.WithMetadata(
						map[string]string{
							"x-envoy-fault-abort-grpc-request": "3",
						},
					),
				),
				testruntime.MustFail,
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "single call update service",
			backendCount: 2,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.AppendEndpointSlices(
					testruntime.BuildEndpointSlices(
						serviceNameV1,
						defaultNamespace,
						backends[0:1],
					),
					testruntime.BuildEndpointSlices(
						serviceNameV2,
						defaultNamespace,
						backends[1:2],
					),
				)
			},
			buildXDSServices: func([]testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildSingleRoute("default"),
						),
						testruntime.WithClusters(
							testruntime.BuildCluster(
								"default",
								testruntime.WithServiceRef(
									gtcv1alpha1.ServiceRef{
										Name: serviceNameV1,
										Port: grpcPort,
									},
								),
							),
						),
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.CallOnce(
				testruntime.BuildCaller(
					testruntime.MethodEcho,
				),
				testruntime.NoCallErrors,
				testruntime.CountByBackendID(
					testruntime.AssertCount("backend-0", 1),
				),
			),
			updateResources: func(t *testing.T, k8s testruntime.FakeK8s, _ []testruntime.Backend) {
				// We update to v2, which means that backend pod should point to a new instance.
				_, err := k8s.GTCApi.ApiV1alpha1().XDSServices("default").Update(
					context.Background(),
					testruntime.Ptr(
						testruntime.BuildXDSService("test-xds",
							"default",
							testruntime.WithRoutes(
								testruntime.BuildSingleRoute("default"),
							),
							testruntime.WithClusters(
								testruntime.BuildCluster(
									"default",
									testruntime.WithServiceRef(
										gtcv1alpha1.ServiceRef{
											Name: serviceNameV2,
											Port: grpcPort,
										},
									),
								),
							),
						),
					),
					metav1.UpdateOptions{},
				)
				require.NoError(t, err)
			},
			doAssertPostUpdate: testruntime.MultiAssert(
				testruntime.Wait(500*time.Millisecond),
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						testruntime.AssertCount("backend-1", 1),
					),
				),
			),
		},
		{
			desc:         "single call update endpointslice",
			backendCount: 2,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.AppendEndpointSlices(
					testruntime.BuildEndpointSlices(
						serviceNameV1,
						defaultNamespace,
						backends[0:1],
					),
				)
			},
			buildXDSServices: func([]testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildSingleRoute("default"),
						),
						testruntime.WithClusters(
							testruntime.BuildCluster(
								"default",
								testruntime.WithServiceRef(
									gtcv1alpha1.ServiceRef{
										Name: serviceNameV1,
										Port: grpcPort,
									},
								),
							),
						),
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.CallOnce(
				testruntime.BuildCaller(
					testruntime.MethodEcho,
				),
				testruntime.NoCallErrors,
				testruntime.CountByBackendID(
					testruntime.AssertCount("backend-0", 1),
				),
			),
			updateResources: func(t *testing.T, k8s testruntime.FakeK8s, backends []testruntime.Backend) {
				// Write the same endpoints, but pointing to the first backend.
				newEps := testruntime.BuildEndpointSlices(serviceNameV1, defaultNamespace, backends[1:2])

				for _, ep := range newEps {
					_, err := k8s.K8s.DiscoveryV1().EndpointSlices(defaultNamespace).Update(
						context.Background(),
						ep.DeepCopy(),
						metav1.UpdateOptions{},
					)
					require.NoError(t, err)
				}
			},
			doAssertPostUpdate: testruntime.MultiAssert(
				// TODO(jly) sadly needed as I don't have any obvious way to wait for the changes to be sent to the client.
				testruntime.Wait(500*time.Millisecond),
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						testruntime.AssertCount("backend-1", 1),
					),
				),
			),
		},
		{
			desc:         "single call update endpointslice with localities",
			backendCount: 2,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.AppendEndpointSlices(
					testruntime.BuildEndpointSlices(
						serviceNameV1,
						defaultNamespace,
						backends[0:1],
					),
				)
			},
			buildXDSServices: func([]testruntime.Backend) []gtcv1alpha1.XDSService {
				return []gtcv1alpha1.XDSService{
					testruntime.BuildXDSService(
						"test-xds",
						"default",
						testruntime.WithRoutes(
							testruntime.BuildSingleRoute("default"),
						),
						testruntime.WithClusters(
							testruntime.BuildCluster(
								"default",
								testruntime.WithLocalities(
									testruntime.BuildLocality(
										testruntime.WithLocalityServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: serviceNameV1,
												Port: grpcPort,
											},
										),
									),
								),
							),
						),
					),
				}
			},
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.CallOnce(
				testruntime.BuildCaller(
					testruntime.MethodEcho,
				),
				testruntime.NoCallErrors,
				testruntime.CountByBackendID(
					testruntime.AssertCount("backend-0", 1),
				),
			),
			updateResources: func(t *testing.T, k8s testruntime.FakeK8s, backends []testruntime.Backend) {
				// Write the same endpoints, but pointing to the first backend.
				newEps := testruntime.BuildEndpointSlices(serviceNameV1, defaultNamespace, backends[1:2])

				for _, ep := range newEps {
					_, err := k8s.K8s.DiscoveryV1().EndpointSlices(defaultNamespace).Update(
						context.Background(),
						ep.DeepCopy(),
						metav1.UpdateOptions{},
					)
					require.NoError(t, err)
				}
			},
			doAssertPostUpdate: testruntime.MultiAssert(
				testruntime.Wait(500*time.Millisecond),
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						testruntime.AssertCount("backend-1", 1),
					),
				),
			),
		},
		{
			desc:         "single call non existing backend",
			backendCount: 2,
			buildEndpointSlices: func(backends []testruntime.Backend) []discoveryv1.EndpointSlice {
				return testruntime.AppendEndpointSlices(
					testruntime.BuildEndpointSlices(
						serviceNameV1,
						defaultNamespace,
						backends[0:1],
					),
				)
			},
			buildXDSServices:    func([]testruntime.Backend) []gtcv1alpha1.XDSService { return nil },
			buildCallContext:    testruntime.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: testruntime.CallOnce(
				testruntime.BuildCaller(
					testruntime.MethodEcho,
					testruntime.WithTimeout(time.Second),
				),
				testruntime.MustFail,
			),
			updateResources: func(t *testing.T, k8s testruntime.FakeK8s, backends []testruntime.Backend) {
				svc := testruntime.BuildXDSService(
					"test-xds",
					"default",
					testruntime.WithRoutes(
						testruntime.BuildSingleRoute("default"),
					),
					testruntime.WithClusters(
						testruntime.BuildCluster(
							"default",
							testruntime.WithServiceRef(
								gtcv1alpha1.ServiceRef{
									Name: serviceNameV1,
									Port: grpcPort,
								},
							),
						),
					),
				)

				_, err := k8s.GTCApi.ApiV1alpha1().XDSServices("default").Create(
					context.Background(),
					&svc,
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
			},
			doAssertPostUpdate: testruntime.MultiAssert(
				testruntime.Wait(500*time.Millisecond),
				testruntime.CallOnce(
					testruntime.BuildCaller(
						testruntime.MethodEcho,
					),
					testruntime.NoCallErrors,
					testruntime.CountByBackendID(
						testruntime.AssertCount("backend-0", 1),
					),
				),
			),
		},
	} {
		t.Run(testCase.desc, func(t *testing.T) {
			backends, err := testruntime.StartBackends(
				testruntime.Config{
					BackendCount: testCase.backendCount,
				},
			)
			require.NoError(t, err)

			defer func() {
				err := backends.Stop()
				require.NoError(t, err)
			}()

			var (
				ctx, cancel = context.WithCancel(context.Background())
				k8s         = testruntime.NewFakeK8s(
					t,
					testCase.buildXDSServices(backends),
					testCase.buildEndpointSlices(backends),
				)
				serverExited = make(chan struct{})
			)

			defer cancel()

			server, err := gtc.NewXDSServer(
				ctx,
				gtc.XDSServerConfig{
					K8sInformers: k8s.K8sInformers,
					GTCInformers: k8s.GTCInformers,
					// TODO(jly): find a way to make this parralelizable.
					// The thing is that having multiple xds servers in parrallel means sadly
					// having multiple values for the XDS_BOOTSTRAP_CONFIG env variables.
					// which is impossible as far as I know.
					BindAddr: ":16000",
				},
				newLogger(t),
			)
			require.NoError(t, err)

			k8s.Start(ctx, t)

			go func() {
				err := server.Run(ctx)
				require.NoError(t, err)
				close(serverExited)
			}()

			// Always explicitely stop the server and wait for it to be finished.
			t.Cleanup(func() {
				cancel()

				<-serverExited
			})

			testCase.setBackendsBehavior(t, backends)

			callCtx := testCase.buildCallContext(t)
			testCase.doAssertPreUpdate(t, callCtx)

			testCase.updateResources(t, k8s, backends)

			testCase.doAssertPostUpdate(t, callCtx)

			err = callCtx.Close()
			require.NoError(t, err)
		})
	}
}

func noChange(*testing.T, testruntime.FakeK8s, []testruntime.Backend) {}
func noAssert(*testing.T, *testruntime.CallContext)                   {}

func answer(t *testing.T, backends testruntime.Backends) {
	backends.SetBehavior(testruntime.DefaultBehavior())
}

func hang(d time.Duration) func(t *testing.T, backends testruntime.Backends) {
	return func(t *testing.T, backends testruntime.Backends) {
		backends.SetBehavior(testruntime.HangBehavior(d))
	}
}

func newLogger(t *testing.T) *zap.Logger {
	if os.Getenv("LOG_LEVEL") == "debug" {
		return zaptest.NewLogger(t)
	}

	return zap.NewNop()
}
