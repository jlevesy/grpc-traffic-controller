package kxds

import (
	"context"
	"strconv"
	"sync/atomic"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	kxdsv1alpha1 "github.com/jlevesy/kxds/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Refresher interface {
	RefreshCache(ctx context.Context, svcs []kxdsv1alpha1.XDSService, endpoints map[ktypes.NamespacedName]corev1.Endpoints) error
}

type cacheRefresher struct {
	xdsCache   cache.SnapshotCache
	hashKey    string
	versionner versionner
}

func NewCacheRefresher(xdsCache cache.SnapshotCache, hashKey string) Refresher {
	return &cacheRefresher{
		xdsCache:   xdsCache,
		hashKey:    hashKey,
		versionner: &atomicIncrementalVersionner{version: 1},
	}
}

func (c *cacheRefresher) RefreshCache(ctx context.Context, svcs []kxdsv1alpha1.XDSService, k8sEndpoints map[ktypes.NamespacedName]corev1.Endpoints) error {
	var (
		listeners []types.Resource
		routes    []types.Resource
		clusters  []types.Resource
		endpoints []types.Resource

		logger = log.FromContext(ctx)
	)

	for _, svc := range svcs {
		eps, ok := k8sEndpoints[ktypes.NamespacedName{Namespace: svc.Namespace, Name: svc.Spec.Destination.Name}]
		if !ok {
			logger.Info(
				"Could not find endpoints, skipping...",
				"ns",
				svc.Namespace,
				"xdsService",
				svc.Name,
				"k8sService",
				svc.Spec.Destination.Name,
			)

			continue
		}

		listenerName := svc.Spec.Listener
		routeName := listenerName + "-route"
		clusterName := listenerName + "-cluster"

		listeners = append(listeners, makeListener(listenerName, routeName))
		routes = append(routes, makeRoute(listenerName, routeName, clusterName))
		clusters = append(clusters, makeCluster(clusterName))
		endpoints = append(endpoints, makeLoadAssignment(clusterName, svc.Spec.Destination.Port, eps))
	}

	version := c.versionner.GetVersion()

	snapshot, err := cache.NewSnapshot(
		version,
		map[resource.Type][]types.Resource{
			resource.ClusterType:  clusters,
			resource.RouteType:    routes,
			resource.ListenerType: listeners,
			resource.EndpointType: endpoints,
		},
	)
	if err != nil {
		logger.Error(err, "Unable to create a new snapshot")
		return err
	}

	logger.Info(
		"Setting a new Snapshot version",
		"version",
		version,
		"listeners",
		len(listeners),
		"routes",
		len(routes),
		"clusters",
		len(clusters),
		"endpoints",
		len(endpoints),
	)

	return c.xdsCache.SetSnapshot(ctx, c.hashKey, snapshot)
}

type versionner interface {
	GetVersion() string
}

type atomicIncrementalVersionner struct {
	version uint64
}

func (v *atomicIncrementalVersionner) GetVersion() string {
	return strconv.FormatUint(atomic.AddUint64(&v.version, 1), 10)
}
