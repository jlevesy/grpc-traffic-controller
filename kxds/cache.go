package kxds

import (
	"context"

	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcesv3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/stream/v3"
	kxdslisters "github.com/jlevesy/kxds/client/listers/kxds/v1alpha1"
	"go.uber.org/zap"
	discoveryv1listers "k8s.io/client-go/listers/discovery/v1"
)

type configWatcher struct {
	resolver     resourceResolver
	watchBuilder watchBuilder

	logger *zap.Logger
}

func newConfigWatcher(endpointSlicesLister discoveryv1listers.EndpointSliceLister, xdsServicesLister kxdslisters.XDSServiceLister, watches watchBuilder, logger *zap.Logger) *configWatcher {
	return &configWatcher{
		logger:       logger.With(zap.String("component", "config_watcher")),
		watchBuilder: watches,
		resolver: resourceTypeResolver{
			resourcesv3.ListenerType: &listenerHandler{xdsServices: xdsServicesLister},
			resourcesv3.ClusterType:  &clusterHandler{xdsServices: xdsServicesLister},
			resourcesv3.EndpointType: &endpointHandler{
				xdsServices:    xdsServicesLister,
				endpointSlices: endpointSlicesLister,
			},
		},
	}
}

func (c *configWatcher) CreateWatch(req *cache.Request, streamState stream.StreamState, resp chan cache.Response) func() {
	ctx, cancelFunc := context.WithCancel(context.Background())

	go c.watch(ctx, streamState, req, resp)

	return cancelFunc
}

func (c *configWatcher) CreateDeltaWatch(*cache.DeltaRequest, stream.StreamState, chan cache.DeltaResponse) func() {
	// This is unsupported.
	c.logger.Error("Received an unexpected CreateDeltaWatch call")
	return nil
}

func (c *configWatcher) watch(ctx context.Context, streamState stream.StreamState, initialReq *cache.Request, respCh chan cache.Response) {
	watch, releaseWatch := c.watchBuilder.buildWatch()
	defer releaseWatch()

	// We're now interested in any update happening about the
	// resources being asked.
	for _, name := range initialReq.ResourceNames {
		watch.watch(resourceRef{
			typeURL:      initialReq.TypeUrl,
			resourceName: name,
		})
	}

	// If we have at least one resource that is not known by the stream
	// state, resend them all.
	// This is required to compute the correct version as it its derived
	// from each kubernetes resource.
	if hasResourceDiff(streamState, initialReq) {
		c.logger.Debug(
			"Sending resources diff",
			zap.String("type", initialReq.TypeUrl),
			zap.Strings("resource_names", initialReq.ResourceNames),
		)

		req := resolveRequest{
			typeUrl:       initialReq.TypeUrl,
			resourceNames: initialReq.ResourceNames,
		}

		if resp, err := c.resolver.resolveResource(req); err == nil {
			sendResponse(ctx, respCh, initialReq, resp)
		}
	}

	for {
		select {
		case <-ctx.Done():
			c.logger.Debug(
				"Exiting watch",
				zap.String("type", initialReq.TypeUrl),
				zap.Strings("resources", initialReq.ResourceNames),
			)
			return
		case ref, ok := <-watch.changes:
			if !ok {
				return
			}

			c.logger.Debug(
				"A resource has changed, sending update",
				zap.String("type", ref.typeURL),
				zap.String("resource", ref.resourceName),
			)

			req := resolveRequest{
				typeUrl:       ref.typeURL,
				resourceNames: []string{ref.resourceName},
			}

			resp, err := c.resolver.resolveResource(req)
			if err != nil {
				c.logger.Error(
					"Unable to resolve resource",
					zap.Error(err),
					zap.String("resource_name", ref.resourceName),
				)
				continue
			}

			sendResponse(ctx, respCh, initialReq, resp)
		}
	}
}

func sendResponse(ctx context.Context, respCh chan cache.Response, req *discoveryv3.DiscoveryRequest, resp *resolveResponse) {
	select {
	case <-ctx.Done():
		return
	case respCh <- &cacheResponse{
		ctx:  ctx,
		req:  req,
		resp: resp,
	}:
	}
}

type cacheResponse struct {
	ctx  context.Context
	req  *discoveryv3.DiscoveryRequest
	resp *resolveResponse
}

func (c *cacheResponse) GetDiscoveryResponse() (*discoveryv3.DiscoveryResponse, error) {
	return &discoveryv3.DiscoveryResponse{
		TypeUrl:     c.resp.typeURL,
		Resources:   c.resp.resources,
		VersionInfo: c.resp.versionInfo,
	}, nil
}

func (c *cacheResponse) GetRequest() *discoveryv3.DiscoveryRequest {
	return c.req
}

func (c *cacheResponse) GetVersion() (string, error) {
	return c.resp.versionInfo, nil
}

func (c *cacheResponse) GetContext() context.Context {
	return c.ctx
}

func hasResourceDiff(streamState stream.StreamState, req *cache.Request) bool {
	knownResourceNames := streamState.GetKnownResourceNames(req.TypeUrl)

	for _, n := range req.ResourceNames {
		if _, ok := knownResourceNames[n]; !ok {
			return true
		}
	}

	return false
}
