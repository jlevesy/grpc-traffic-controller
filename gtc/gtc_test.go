package gtc_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/xds"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gtcv1alpha1 "github.com/jlevesy/grpc-traffic-controller/api/gtc/v1alpha1"
	"github.com/jlevesy/grpc-traffic-controller/gtc"
	tr "github.com/jlevesy/grpc-traffic-controller/pkg/testruntime"
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
)

func TestServer(t *testing.T) {
	for _, testCase := range []struct {
		desc                string
		backendCount        int
		buildEndpointSlices func(backends []tr.Backend) []discoveryv1.EndpointSlice
		buildGRPCListeners  func(backends []tr.Backend) []gtcv1alpha1.GRPCListener
		buildCallContext    func(t *testing.T) *tr.CallContext
		setBackendsBehavior func(t *testing.T, bs tr.Backends)
		doAssertPreUpdate   func(t *testing.T, callCtx *tr.CallContext)
		updateResources     func(t *testing.T, k8s tr.FakeK8s, backends []tr.Backend)
		doAssertPostUpdate  func(t *testing.T, callCtx *tr.CallContext)
	}{
		{
			desc:         "single call port by name",
			backendCount: 1,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.BuildEndpointSlices(
					serviceNameV1,
					defaultNamespace,
					backends[0:1],
				)
			},
			buildGRPCListeners: func([]tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.CallOnce(
				tr.BuildCaller(
					tr.MethodEcho,
				),
				tr.NoCallErrors,
				tr.CountByBackendID(
					tr.AssertCount("backend-0", 1),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "single call port by number",
			backendCount: 1,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.BuildEndpointSlices(
					serviceNameV1,
					defaultNamespace,
					backends[0:1],
				)
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
						),
					),
				}
			},
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.CallOnce(
				tr.BuildCaller(
					tr.MethodEcho,
				),
				tr.NoCallErrors,
				tr.CountByBackendID(
					tr.AssertCount("backend-0", 1),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "cross namespace",
			backendCount: 1,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.BuildEndpointSlices(
					serviceNameV1,
					"some-app",
					backends[0:1],
				)
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
											gtcv1alpha1.ServiceRef{
												Name:      serviceNameV1,
												Namespace: "some-app",
												Port:      grpcPort,
											},
										),
									),
								),
							),
						),
					),
				}
			},
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.CallOnce(
				tr.BuildCaller(
					tr.MethodEcho,
				),
				tr.NoCallErrors,
				tr.CountByBackendID(
					tr.AssertCount("backend-0", 1),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "locality based wrr",
			backendCount: 4,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.AppendEndpointSlices(
					tr.BuildEndpointSlices("test-service", "default", backends[0:2]),
					tr.BuildEndpointSlices("test-service-v2", "default", backends[2:4]),
				)
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithLocalities(
											tr.BuildLocality(
												tr.WithLocalityWeight(80),
												tr.WithLocalityServiceRef(
													gtcv1alpha1.ServiceRef{
														Name: "test-service",
														Port: grpcPort,
													},
												),
											),
											tr.BuildLocality(
												tr.WithLocalityWeight(20),
												tr.WithLocalityServiceRef(
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
						),
					),
				}
			},
			setBackendsBehavior: answer,
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			doAssertPreUpdate: tr.CallN(
				tr.BuildCaller(
					tr.MethodEcho,
				),
				10000,
				tr.NoCallErrors,
				tr.CountByBackendID(
					// 80% of calls
					tr.AssertCountWithinDelta("backend-0", 4000, 500.0),
					tr.AssertCountWithinDelta("backend-1", 4000, 500.0),
					// 20% of calls
					tr.AssertCountWithinDelta("backend-2", 1000, 500.0),
					tr.AssertCountWithinDelta("backend-3", 1000, 500.0),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "priority fallback",
			backendCount: 2,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.AppendEndpointSlices(
					// No backend for this service.
					tr.BuildEndpointSlices("test-service", "default", backends[0:0]),
					tr.BuildEndpointSlices("test-service-v2", "default", backends[1:2]),
				)
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithLocalities(
											tr.BuildLocality(
												tr.WithLocalityPriority(0),
												tr.WithLocalityServiceRef(
													gtcv1alpha1.ServiceRef{
														Name: "test-service",
														Port: grpcPort,
													},
												),
											),
											tr.BuildLocality(
												tr.WithLocalityPriority(1),
												tr.WithLocalityServiceRef(
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
						),
					),
				}
			},
			setBackendsBehavior: answer,
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			doAssertPreUpdate: tr.CallOnce(
				tr.BuildCaller(
					tr.MethodEcho,
				),
				tr.NoCallErrors,
				tr.CountByBackendID(
					// No calls for the first set of backends
					tr.AssertCount("backend-0", 0),
					// One call for the second backend.
					tr.AssertCount("backend-1", 1),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "route matcher method matching",
			backendCount: 2,

			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.AppendEndpointSlices(
					tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					tr.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithRouteMatcher(
									tr.BuildRouteMatcher(
										tr.WithMethodMatcher(
											"echo",
											"Echo",
											"EchoPremium",
										),
									),
								),
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: serviceNameV2,
												Port: grpcPort,
											},
										),
									),
								),
							),
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.MultiAssert(
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEchoPremium,
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						// One call for the second backend, because we're calling premium.
						tr.AssertCount("backend-1", 1),
						tr.AssertCount("backend-0", 0),
					),
				),
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						// No calls for the first set of backends
						// First backend should get a call.
						tr.AssertCount("backend-0", 1),
						tr.AssertCount("backend-1", 0),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "route matcher service matcher",
			backendCount: 2,

			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.AppendEndpointSlices(
					tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					tr.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithRouteMatcher(
									tr.BuildRouteMatcher(
										tr.WithServiceMatcher(
											"echo",
											"Echo",
										),
									),
								),
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: serviceNameV2,
												Port: grpcPort,
											},
										),
									),
								),
							),
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.CallOnce(
				tr.BuildCaller(
					tr.MethodEcho,
				),
				tr.NoCallErrors,
				tr.CountByBackendID(
					// One call for the second backend, because we're calling premium.
					tr.AssertCount("backend-1", 1),
					tr.AssertCount("backend-0", 0),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "route matcher namespace matcher",
			backendCount: 2,

			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.AppendEndpointSlices(
					tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					tr.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithRouteMatcher(
									tr.BuildRouteMatcher(
										tr.WithNamespaceMatcher("echo"),
									),
								),
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: serviceNameV2,
												Port: grpcPort,
											},
										),
									),
								),
							),
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.CallOnce(
				tr.BuildCaller(
					tr.MethodEcho,
				),
				tr.NoCallErrors,
				tr.CountByBackendID(
					// One call for the second backend, because we're calling premium.
					tr.AssertCount("backend-1", 1),
					tr.AssertCount("backend-0", 0),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "route matcher metadata invert matching",
			backendCount: 2,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.AppendEndpointSlices(
					tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					tr.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithRouteMatcher(
									tr.BuildRouteMatcher(
										tr.WithMetadataMatchers(
											tr.MetadataInvertMatch(
												tr.MetadataExactMatch(
													"x-variant",
													"Awesome",
												),
											),
										),
									),
								),
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: serviceNameV2,
												Port: grpcPort,
											},
										),
									),
								),
							),
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.MultiAssert(
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
						tr.WithMetadata(
							map[string]string{
								"x-variant": "Awesome",
							},
						),
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						tr.AssertCount("backend-1", 0),
						tr.AssertCount("backend-0", 1),
					),
				),
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
						tr.WithMetadata(
							map[string]string{
								"x-variant": "NotAwesome",
							},
						),
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						tr.AssertCount("backend-0", 0),
						tr.AssertCount("backend-1", 1),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "route matcher metadata exact matching",
			backendCount: 2,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.AppendEndpointSlices(
					tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					tr.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithRouteMatcher(
									tr.BuildRouteMatcher(
										tr.WithMetadataMatchers(
											tr.MetadataExactMatch(
												"x-variant",
												"Awesome",
											),
										),
									),
								),
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: serviceNameV2,
												Port: grpcPort,
											},
										),
									),
								),
							),
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.MultiAssert(
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
						tr.WithMetadata(
							map[string]string{
								"x-variant": "Awesome",
							},
						),
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						// One call for the second backend, because we're calling premium.
						tr.AssertCount("backend-1", 1),
						tr.AssertCount("backend-0", 0),
					),
				),
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						// No calls for the first set of backends
						// First backend should get a call.
						tr.AssertCount("backend-0", 1),
						tr.AssertCount("backend-1", 0),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "route matcher metadata safe regex match",
			backendCount: 2,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.AppendEndpointSlices(
					tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					tr.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithRouteMatcher(
									tr.BuildRouteMatcher(
										tr.WithMetadataMatchers(
											gtcv1alpha1.MetadataMatcher{
												Name: "x-variant",
												Regex: &gtcv1alpha1.RegexMatcher{
													Regex:  "Awe.*",
													Engine: "re2",
												},
											},
										),
									),
								),
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: serviceNameV2,
												Port: grpcPort,
											},
										),
									),
								),
							),
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.MultiAssert(
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
						tr.WithMetadata(
							map[string]string{
								"x-variant": "Awesome",
							},
						),
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						// One call for the second backend, because we're calling premium.
						tr.AssertCount("backend-1", 1),
						tr.AssertCount("backend-0", 0),
					),
				),
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						// No calls for the first set of backends
						// First backend should get a call.
						tr.AssertCount("backend-0", 1),
						tr.AssertCount("backend-1", 0),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "route matcher metadata range match",
			backendCount: 2,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.AppendEndpointSlices(
					tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					tr.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithRouteMatcher(
									tr.BuildRouteMatcher(
										tr.WithMetadataMatchers(
											gtcv1alpha1.MetadataMatcher{
												Name: "x-variant",
												Range: &gtcv1alpha1.RangeMatcher{
													Start: 10,
													End:   20,
												},
											},
										),
									),
								),
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: serviceNameV2,
												Port: grpcPort,
											},
										),
									),
								),
							),
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.MultiAssert(
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
						tr.WithMetadata(
							map[string]string{
								// In range, call backend 1.
								"x-variant": "12",
							},
						),
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						tr.AssertCount("backend-1", 1),
						tr.AssertCount("backend-0", 0),
					),
				),
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
						tr.WithMetadata(
							map[string]string{
								// Out of bound, call backend-0.
								"x-variant": "9",
							},
						),
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						tr.AssertCount("backend-0", 1),
						tr.AssertCount("backend-1", 0),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "route matcher metadata present match",
			backendCount: 2,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.AppendEndpointSlices(
					tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					tr.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithRouteMatcher(
									tr.BuildRouteMatcher(
										tr.WithMetadataMatchers(
											tr.MetadataPresentMatch(
												"x-variant",
												true,
											),
										),
									),
								),
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: serviceNameV2,
												Port: grpcPort,
											},
										),
									),
								),
							),
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.MultiAssert(
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
						tr.WithMetadata(
							map[string]string{
								// Header is present, send to v2.
								"x-variant": "wooop",
							},
						),
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						tr.AssertCount("backend-1", 1),
						tr.AssertCount("backend-0", 0),
					),
				),
				tr.CallOnce(
					tr.BuildCaller(tr.MethodEcho),
					tr.NoCallErrors,
					tr.CountByBackendID(
						tr.AssertCount("backend-0", 1),
						tr.AssertCount("backend-1", 0),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "route matcher metadata prefix match",
			backendCount: 2,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.AppendEndpointSlices(
					tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					tr.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithRouteMatcher(
									tr.BuildRouteMatcher(
										tr.WithMetadataMatchers(
											tr.MetadataPrefixMatch(
												"x-variant",
												"wo",
											),
										),
									),
								),
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: serviceNameV2,
												Port: grpcPort,
											},
										),
									),
								),
							),
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.MultiAssert(
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
						tr.WithMetadata(
							map[string]string{
								// Header has the prefix wo, send to v2.
								"x-variant": "wooop",
							},
						),
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						tr.AssertCount("backend-1", 1),
						tr.AssertCount("backend-0", 0),
					),
				),
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
						tr.WithMetadata(
							map[string]string{
								// Header has not the prefix wo, send to v1.
								"x-variant": "not",
							},
						),
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						tr.AssertCount("backend-0", 1),
						tr.AssertCount("backend-1", 0),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "route matcher metadata suffix match",
			backendCount: 2,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.AppendEndpointSlices(
					tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					tr.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithRouteMatcher(
									tr.BuildRouteMatcher(
										tr.WithMetadataMatchers(
											tr.MetadataSuffixMatch(
												"x-variant",
												"oop",
											),
										),
									),
								),
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: serviceNameV2,
												Port: grpcPort,
											},
										),
									),
								),
							),
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.MultiAssert(
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
						tr.WithMetadata(
							map[string]string{
								// Header has the sufix oop, send to v2.
								"x-variant": "wooop",
							},
						),
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						tr.AssertCount("backend-1", 1),
						tr.AssertCount("backend-0", 0),
					),
				),
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
						tr.WithMetadata(
							map[string]string{
								// Header has not the suffix oop, send to v1.
								"x-variant": "not",
							},
						),
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						tr.AssertCount("backend-0", 1),
						tr.AssertCount("backend-1", 0),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "route matcher runtime fraction traffic splitting",
			backendCount: 2,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.AppendEndpointSlices(
					tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1]),
					tr.BuildEndpointSlices(serviceNameV2, "default", backends[1:2]),
				)
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithRouteMatcher(
									tr.BuildRouteMatcher(
										tr.WithFractionMatcher(
											gtcv1alpha1.Fraction{
												Numerator:   20,
												Denominator: "hundred",
											},
										),
									),
								),
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: serviceNameV2,
												Port: grpcPort,
											},
										),
									),
								),
							),
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.CallN(
				tr.BuildCaller(
					tr.MethodEcho,
				),
				10000,
				tr.NoCallErrors,
				tr.CountByBackendID(
					tr.AssertCountWithinDelta("backend-1", 2000, 500.0),
					tr.AssertCountWithinDelta("backend-0", 8000, 500.0),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "listener max stream duration",
			backendCount: 1,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1])
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithMaxStreamDuration(50*time.Millisecond),
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: hang(10 * time.Second),
			doAssertPreUpdate: tr.WithinDelay(
				time.Second,
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
					),
					tr.MustFail,
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "route max stream duration",
			backendCount: 1,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1])
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithRouteMaxStreamDuration(50*time.Millisecond),
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: hang(10 * time.Second),
			doAssertPreUpdate: tr.WithinDelay(
				time.Second,
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
					),
					tr.MustFail,
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "max requests on backend",
			backendCount: 1,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1])
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithMaxRequests(1),
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: hang(1 * time.Second),
			doAssertPreUpdate: tr.CallNParallel(
				tr.BuildCaller(
					tr.MethodEcho,
				),
				10,
				tr.AggregateByError(
					tr.AssertCount("ok", 1),
					tr.AssertAggregatedValuePartial(
						"rpc error: code = Unavailable desc = max requests 1 exceeded",
						9,
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "listener interceptors fixed delay injection",
			backendCount: 1,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1])
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: serviceNameV1,
												Port: grpcPort,
											},
										),
									),
								),
							),
						),
						tr.WithInterceptors(
							gtcv1alpha1.Interceptor{
								Fault: &gtcv1alpha1.FaultInterceptor{
									Delay: &gtcv1alpha1.FaultDelay{
										Fixed: tr.DurationPtr(500 * time.Millisecond),
										Percentage: &gtcv1alpha1.Fraction{
											Numerator:   100,
											Denominator: "hundred",
										},
									},
								},
							},
						),
					),
				}
			},
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.ExceedDelay(
				200*time.Millisecond,
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
					),
					tr.NoCallErrors,
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "listener interceptors delay injection",
			backendCount: 1,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1])
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: serviceNameV1,
												Port: grpcPort,
											},
										),
									),
								),
							),
						),
						tr.WithInterceptors(
							gtcv1alpha1.Interceptor{
								Fault: &gtcv1alpha1.FaultInterceptor{
									Delay: &gtcv1alpha1.FaultDelay{
										Metadata: &gtcv1alpha1.MetadataFault{},
										Percentage: &gtcv1alpha1.Fraction{
											Numerator:   100,
											Denominator: "hundred",
										},
									},
								},
							},
						),
					),
				}
			},
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.ExceedDelay(
				200*time.Millisecond,
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
						tr.WithMetadata(
							map[string]string{
								// Delay by 500ms.
								"x-envoy-fault-delay-request": "500",
							},
						),
					),
					tr.NoCallErrors,
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "listener interceptors abort injection grpc",
			backendCount: 1,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1])
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: serviceNameV1,
												Port: grpcPort,
											},
										),
									),
								),
							),
						),
						tr.WithInterceptors(
							gtcv1alpha1.Interceptor{
								Fault: &gtcv1alpha1.FaultInterceptor{
									Abort: &gtcv1alpha1.FaultAbort{
										Code: tr.Ptr(uint32(4)),
										Percentage: &gtcv1alpha1.Fraction{
											Numerator:   100,
											Denominator: "hundred",
										},
									},
								},
							},
						),
					),
				}
			},
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.CallOnce(
				tr.BuildCaller(
					tr.MethodEcho,
				),
				tr.MustFail,
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "listerner interceptors abort metadata grpc",
			backendCount: 1,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1])
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: serviceNameV1,
												Port: grpcPort,
											},
										),
									),
								),
							),
						),
						tr.WithInterceptors(
							gtcv1alpha1.Interceptor{
								Fault: &gtcv1alpha1.FaultInterceptor{
									Abort: &gtcv1alpha1.FaultAbort{
										Metadata: &gtcv1alpha1.MetadataFault{},
										Percentage: &gtcv1alpha1.Fraction{
											Numerator:   100,
											Denominator: "hundred",
										},
									},
								},
							},
						),
					),
				}
			},
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.CallOnce(
				tr.BuildCaller(
					tr.MethodEcho,
					tr.WithMetadata(
						map[string]string{
							"x-envoy-fault-abort-grpc-request": "3",
						},
					),
				),
				tr.MustFail,
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "route interceptors overrides abort",
			backendCount: 1,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1])
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithInterceptors(
							gtcv1alpha1.Interceptor{
								Fault: &gtcv1alpha1.FaultInterceptor{
									Abort: &gtcv1alpha1.FaultAbort{
										Code: tr.Ptr(uint32(10)),
										Percentage: &gtcv1alpha1.Fraction{
											Numerator:   100,
											Denominator: "hundred",
										},
									},
								},
							},
						),
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithRouteMatcher(
									tr.BuildRouteMatcher(
										tr.WithMethodMatcher("echo", "Echo", "EchoPremium"),
									),
								),
								tr.WithRouteInterceptorOverrides(
									gtcv1alpha1.Interceptor{
										Fault: &gtcv1alpha1.FaultInterceptor{
											Abort: &gtcv1alpha1.FaultAbort{
												Code: tr.Ptr(uint32(15)),
												Percentage: &gtcv1alpha1.Fraction{
													Numerator:   100,
													Denominator: "hundred",
												},
											},
										},
									},
								),
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: serviceNameV2,
												Port: grpcPort,
											},
										),
									),
								),
							),
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.MultiAssert(
				tr.CallOnce(
					tr.BuildCaller(tr.MethodEchoPremium),
					tr.MustFailWithCode(codes.DataLoss),
				),
				tr.CallOnce(
					tr.BuildCaller(tr.MethodEcho),
					tr.MustFailWithCode(codes.Aborted),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "backend interceptors overrides abort",
			backendCount: 1,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.BuildEndpointSlices(serviceNameV1, "default", backends[0:1])
			},
			buildGRPCListeners: func(backends []tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithInterceptors(
							gtcv1alpha1.Interceptor{
								Fault: &gtcv1alpha1.FaultInterceptor{
									Abort: &gtcv1alpha1.FaultAbort{
										Code: tr.Ptr(uint32(10)),
										Percentage: &gtcv1alpha1.Fraction{
											Numerator:   100,
											Denominator: "hundred",
										},
									},
								},
							},
						),
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithRouteMatcher(
									tr.BuildRouteMatcher(
										tr.WithMethodMatcher("echo", "Echo", "EchoPremium"),
									),
								),
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithBackendInterceptorOverrides(
											gtcv1alpha1.Interceptor{
												Fault: &gtcv1alpha1.FaultInterceptor{
													Abort: &gtcv1alpha1.FaultAbort{
														Code: tr.Ptr(uint32(15)),
														Percentage: &gtcv1alpha1.Fraction{
															Numerator:   100,
															Denominator: "hundred",
														},
													},
												},
											},
										),
										tr.WithServiceRef(
											gtcv1alpha1.ServiceRef{
												Name: serviceNameV2,
												Port: grpcPort,
											},
										),
									),
								),
							),
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.MultiAssert(
				tr.CallOnce(
					tr.BuildCaller(tr.MethodEchoPremium),
					tr.MustFailWithCode(codes.DataLoss),
				),
				tr.CallOnce(
					tr.BuildCaller(tr.MethodEcho),
					tr.MustFailWithCode(codes.Aborted),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "update single call service",
			backendCount: 2,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.AppendEndpointSlices(
					tr.BuildEndpointSlices(
						serviceNameV1,
						defaultNamespace,
						backends[0:1],
					),
					tr.BuildEndpointSlices(
						serviceNameV2,
						defaultNamespace,
						backends[1:2],
					),
				)
			},
			buildGRPCListeners: func([]tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.CallOnce(
				tr.BuildCaller(
					tr.MethodEcho,
				),
				tr.NoCallErrors,
				tr.CountByBackendID(
					tr.AssertCount("backend-0", 1),
				),
			),
			updateResources: func(t *testing.T, k8s tr.FakeK8s, _ []tr.Backend) {
				// We switch the backend to v2, which means that backend pod should point to a new instance.
				_, err := k8s.GTCApi.ApiV1alpha1().GRPCListeners("default").Update(
					context.Background(),
					tr.Ptr(
						tr.BuildGRPCListener(
							"test-xds",
							"default",
							tr.WithRoutes(
								tr.BuildRoute(
									tr.WithBackends(
										tr.BuildBackend(
											tr.WithServiceRef(
												gtcv1alpha1.ServiceRef{
													Name: serviceNameV2,
													Port: grpcPort,
												},
											),
										),
									),
								),
							),
						),
					),
					metav1.UpdateOptions{},
				)
				require.NoError(t, err)
			},
			doAssertPostUpdate: tr.MultiAssert(
				tr.Wait(500*time.Millisecond),
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						tr.AssertCount("backend-1", 1),
					),
				),
			),
		},
		{
			desc:         "update single call endpointslice",
			backendCount: 2,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.AppendEndpointSlices(
					tr.BuildEndpointSlices(
						serviceNameV1,
						defaultNamespace,
						backends[0:1],
					),
				)
			},
			buildGRPCListeners: func([]tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.CallOnce(
				tr.BuildCaller(
					tr.MethodEcho,
				),
				tr.NoCallErrors,
				tr.CountByBackendID(
					tr.AssertCount("backend-0", 1),
				),
			),
			updateResources: func(t *testing.T, k8s tr.FakeK8s, backends []tr.Backend) {
				// Write the same endpoints, but pointing to the backend at index 1.
				newEps := tr.BuildEndpointSlices(serviceNameV1, defaultNamespace, backends[1:2])

				for _, ep := range newEps {
					_, err := k8s.K8s.DiscoveryV1().EndpointSlices(defaultNamespace).Update(
						context.Background(),
						ep.DeepCopy(),
						metav1.UpdateOptions{},
					)
					require.NoError(t, err)
				}
			},
			doAssertPostUpdate: tr.MultiAssert(
				tr.Wait(500*time.Millisecond),
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						tr.AssertCount("backend-1", 1),
					),
				),
			),
		},
		{
			desc:         "update endpointslice with localities",
			backendCount: 2,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.AppendEndpointSlices(
					tr.BuildEndpointSlices(
						serviceNameV1,
						defaultNamespace,
						backends[0:1],
					),
				)
			},
			buildGRPCListeners: func([]tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithLocalities(
											tr.BuildLocality(
												tr.WithLocalityServiceRef(
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
						),
					),
				}
			},
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.CallOnce(
				tr.BuildCaller(
					tr.MethodEcho,
				),
				tr.NoCallErrors,
				tr.CountByBackendID(
					tr.AssertCount("backend-0", 1),
				),
			),
			updateResources: func(t *testing.T, k8s tr.FakeK8s, backends []tr.Backend) {
				// Write the same endpoints, but pointing to the first backend.
				newEps := tr.BuildEndpointSlices(serviceNameV1, defaultNamespace, backends[1:2])

				for _, ep := range newEps {
					_, err := k8s.K8s.DiscoveryV1().EndpointSlices(defaultNamespace).Update(
						context.Background(),
						ep.DeepCopy(),
						metav1.UpdateOptions{},
					)
					require.NoError(t, err)
				}
			},
			doAssertPostUpdate: tr.MultiAssert(
				tr.Wait(500*time.Millisecond),
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						tr.AssertCount("backend-1", 1),
					),
				),
			),
		},
		{
			desc:         "update single call non existing backend",
			backendCount: 2,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.AppendEndpointSlices(
					tr.BuildEndpointSlices(
						serviceNameV1,
						defaultNamespace,
						backends[0:1],
					),
				)
			},
			buildGRPCListeners:  func([]tr.Backend) []gtcv1alpha1.GRPCListener { return nil },
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.CallOnce(
				tr.BuildCaller(
					tr.MethodEcho,
					tr.WithTimeout(time.Second),
				),
				tr.MustFail,
			),
			updateResources: func(t *testing.T, k8s tr.FakeK8s, backends []tr.Backend) {
				lis := tr.BuildGRPCListener(
					"test-xds",
					"default",
					tr.WithRoutes(
						tr.BuildRoute(
							tr.WithBackends(
								tr.BuildBackend(
									tr.WithServiceRef(
										gtcv1alpha1.ServiceRef{
											Name: serviceNameV1,
											Port: grpcPort,
										},
									),
								),
							),
						),
					),
				)

				_, err := k8s.GTCApi.ApiV1alpha1().GRPCListeners("default").Create(
					context.Background(),
					&lis,
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
			},
			doAssertPostUpdate: tr.MultiAssert(
				tr.Wait(500*time.Millisecond),
				tr.CallOnce(
					tr.BuildCaller(
						tr.MethodEcho,
					),
					tr.NoCallErrors,
					tr.CountByBackendID(
						tr.AssertCount("backend-0", 1),
					),
				),
			),
		},
		{
			desc:         "listener retry",
			backendCount: 1,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.BuildEndpointSlices(
					serviceNameV1,
					defaultNamespace,
					backends[0:1],
				)
			},
			buildGRPCListeners: func([]tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithListenerRetry(
							gtcv1alpha1.RetryPolicy{
								RetryOn: []string{
									"unavailable",
									"cancelled",
								},
								NumRetries: tr.Ptr(uint32(2)),
								Backoff: &gtcv1alpha1.RetryBackoff{
									BaseInterval: metav1.Duration{
										Duration: 500 * time.Millisecond,
									},
									MaxInterval: tr.DurationPtr(time.Second),
								},
							},
						),
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: sequence(codes.Unavailable, codes.Canceled),
			doAssertPreUpdate: tr.CallOnce(
				tr.BuildCaller(
					tr.MethodEcho,
				),
				tr.NoCallErrors,
				tr.CountByBackendID(
					tr.AssertCount("backend-0", 1),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "route retry",
			backendCount: 1,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.BuildEndpointSlices(
					serviceNameV1,
					defaultNamespace,
					backends[0:1],
				)
			},
			buildGRPCListeners: func([]tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithRouteRetry(
									gtcv1alpha1.RetryPolicy{
										RetryOn: []string{
											"unavailable",
											"cancelled",
										},
										NumRetries: tr.Ptr(uint32(2)),
										Backoff: &gtcv1alpha1.RetryBackoff{
											BaseInterval: metav1.Duration{
												Duration: 500 * time.Millisecond,
											},
											MaxInterval: tr.DurationPtr(time.Second),
										},
									},
								),
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: sequence(codes.Canceled, codes.Unavailable),
			doAssertPreUpdate: tr.CallOnce(
				tr.BuildCaller(
					tr.MethodEcho,
				),
				tr.NoCallErrors,
				tr.CountByBackendID(
					tr.AssertCount("backend-0", 1),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
		{
			desc:         "ring hash backend lb policy",
			backendCount: 2,
			buildEndpointSlices: func(backends []tr.Backend) []discoveryv1.EndpointSlice {
				return tr.BuildEndpointSlices(
					serviceNameV1,
					defaultNamespace,
					backends,
				)
			},
			buildGRPCListeners: func([]tr.Backend) []gtcv1alpha1.GRPCListener {
				return []gtcv1alpha1.GRPCListener{
					tr.BuildGRPCListener(
						"test-xds",
						"default",
						tr.WithRoutes(
							tr.BuildRoute(
								tr.WithRouteHashPolicy(
									gtcv1alpha1.HashPolicy{Metadata: "country"},
								),
								tr.WithBackends(
									tr.BuildBackend(
										tr.WithBackendLBPolicy("ring_hash"),
										tr.WithBackendRingHashConfig(
											gtcv1alpha1.RingHashConfig{
												MinRingSize: 1024,
												MaxRingSize: 838860,
											},
										),
										tr.WithServiceRef(
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
			buildCallContext:    tr.DefaultCallContext("xds:///default/test-xds"),
			setBackendsBehavior: answer,
			doAssertPreUpdate: tr.MultiAssert(
				tr.CallN(
					tr.BuildCaller(
						tr.MethodEcho,
						tr.WithMetadata(map[string]string{"country": "france"}),
					),
					10,
					tr.NoCallErrors,
					tr.CountByBackendID(
						tr.AssertOneBackendGotAllCalls(10),
					),
				),
				tr.CallN(
					tr.BuildCaller(
						tr.MethodEcho,
						tr.WithMetadata(map[string]string{"country": "sweden"}),
					),
					10,
					tr.NoCallErrors,
					tr.CountByBackendID(
						tr.AssertOneBackendGotAllCalls(10),
					),
				),
			),
			updateResources:    noChange,
			doAssertPostUpdate: noAssert,
		},
	} {
		t.Run(testCase.desc, func(t *testing.T) {
			backends, err := tr.StartBackends(
				tr.Config{
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
				k8s         = tr.NewFakeK8s(
					t,
					testCase.buildGRPCListeners(backends),
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

func noChange(*testing.T, tr.FakeK8s, []tr.Backend) {}
func noAssert(*testing.T, *tr.CallContext)          {}

func answer(t *testing.T, backends tr.Backends) {
	backends.SetBehavior(tr.DefaultBehavior())
}

func hang(d time.Duration) func(t *testing.T, backends tr.Backends) {
	return func(t *testing.T, backends tr.Backends) {
		backends.SetBehavior(tr.HangBehavior(d))
	}
}

func sequence(cs ...codes.Code) func(t *testing.T, backends tr.Backends) {
	return func(t *testing.T, backends tr.Backends) {
		backends.SetBehavior(tr.SequenceBehavior(cs...))
	}
}

func newLogger(t *testing.T) *zap.Logger {
	if os.Getenv("LOG_LEVEL") == "debug" {
		return zaptest.NewLogger(t)
	}

	return zap.NewNop()
}
