package kxds

import (
	"fmt"
	"strings"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	resourcesv3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	anyv1 "github.com/golang/protobuf/ptypes/any"
	kxdslisters "github.com/jlevesy/grpc-traffic-controller/client/listers/kxds/v1alpha1"
)

type listenerHandler struct {
	xdsServices kxdslisters.XDSServiceLister
}

func (h *listenerHandler) resolveResource(req resolveRequest) (*resolveResponse, error) {
	var (
		err      error
		versions = make([]string, len(req.resourceNames))
		response = resolveResponse{
			typeURL:   resourcesv3.ListenerType,
			resources: make([]*anyv1.Any, len(req.resourceNames)),
		}
	)

	for i, resourceName := range req.resourceNames {
		resource, version, err := h.makeListener(resourceName)
		if err != nil {
			return nil, err
		}

		response.resources[i], err = encodeResource(req.typeUrl, resource)
		if err != nil {
			return nil, err
		}

		versions[i] = version
	}

	response.versionInfo, err = computeVersionInfo(versions)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (h *listenerHandler) makeListener(resourceName string) (*listenerv3.Listener, string, error) {
	namespace, name, err := parseListenerName(resourceName)
	if err != nil {
		return nil, "", err
	}

	svc, err := h.xdsServices.XDSServices(namespace).Get(name)
	if err != nil {
		return nil, "", err
	}

	filters, err := makeFilters(svc.Spec.Filters)
	if err != nil {
		return nil, "", err
	}

	routeConfig, err := makeRouteConfig(resourceName, svc)
	if err != nil {
		return nil, "", err
	}

	httpConnManager := &hcm.HttpConnectionManager{
		CommonHttpProtocolOptions: &core.HttpProtocolOptions{
			MaxStreamDuration: makeDuration(svc.Spec.MaxStreamDuration),
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
	}, svc.ResourceVersion, nil
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
