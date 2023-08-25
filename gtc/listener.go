package gtc

import (
	"fmt"
	"strings"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	resourcesv3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	gtclisters "github.com/jlevesy/grpc-traffic-controller/client/listers/gtc/v1alpha1"
)

type listenerHandler struct {
	grpcListeners gtclisters.GRPCListenerLister
}

func (h *listenerHandler) resolveResource(req resolveRequest) (*resolveResponse, error) {
	response := newResolveResponse(resourcesv3.ListenerType, len(req.resourceNames))

	for i, resourceName := range req.resourceNames {
		resource, version, err := h.makeListener(resourceName)
		if err != nil {
			return nil, err
		}

		response.resources[i], err = encodeResource(req.typeUrl, resource)
		if err != nil {
			return nil, err
		}

		if err := response.useResourceVersion(version); err != nil {
			return nil, err
		}
	}

	return response, nil
}

func (h *listenerHandler) makeListener(resourceName string) (*listenerv3.Listener, string, error) {
	namespace, name, err := parseListenerName(resourceName)
	if err != nil {
		return nil, "", err
	}

	listener, err := h.grpcListeners.GRPCListeners(namespace).Get(name)
	if err != nil {
		return nil, "", err
	}

	filters, err := makeFilters(listener.Spec.Interceptors)
	if err != nil {
		return nil, "", err
	}

	routeConfig, err := makeRouteConfig(resourceName, listener)
	if err != nil {
		return nil, "", err
	}

	httpConnManager := &hcm.HttpConnectionManager{
		CommonHttpProtocolOptions: &core.HttpProtocolOptions{
			MaxStreamDuration: makeDuration(listener.Spec.MaxStreamDuration),
		},
		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: routeConfig,
		},
		HttpFilters: filters,
	}

	return &listenerv3.Listener{
		Name: resourceName,
		ApiListener: &listenerv3.ApiListener{
			ApiListener: mustAny(httpConnManager),
		},
	}, listener.ResourceVersion, nil
}

func parseListenerName(resourceName string) (string, string, error) {
	sp := strings.SplitN(resourceName, "/", 2)
	if len(sp) != 2 {
		return "", "", malformedListenerResourceNameError(resourceName)
	}

	return sp[0], sp[1], nil
}

func listenerName(namespace, name string) string {
	return namespace + "/" + name
}

type malformedListenerResourceNameError string

func (m malformedListenerResourceNameError) Error() string {
	return fmt.Sprintf("could not parse listener resource name %s", string(m))
}
