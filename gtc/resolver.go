package gtc

import (
	"fmt"

	anyv1 "github.com/golang/protobuf/ptypes/any"
)

type resolveRequest struct {
	resourceNames []string
	typeUrl       string
}

type resolveResponse struct {
	typeURL     string
	resources   []*anyv1.Any
	versionInfo string
}

type resourceResolver interface {
	resolveResource(req resolveRequest) (*resolveResponse, error)
}

type resourceTypeResolver map[string]resourceResolver

func (h resourceTypeResolver) resolveResource(req resolveRequest) (*resolveResponse, error) {
	hdl, ok := h[req.typeUrl]
	if !ok {
		return nil, unsupportedResourceType(req.typeUrl)
	}

	return hdl.resolveResource(req)
}

type unsupportedResourceType string

func (u unsupportedResourceType) Error() string {
	return fmt.Sprintf("Unsupported type URL %q", string(u))
}
