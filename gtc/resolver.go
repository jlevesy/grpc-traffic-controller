package gtc

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"

	anyv1 "github.com/golang/protobuf/ptypes/any"
)

type resolveRequest struct {
	resourceNames []string
	typeUrl       string
}

type resolveResponse struct {
	typeURL       string
	resources     []*anyv1.Any
	versionHasher hash.Hash
}

func newResolveResponse(typeURL string, resourcesCount int) *resolveResponse {
	return &resolveResponse{
		typeURL:       typeURL,
		resources:     make([]*anyv1.Any, resourcesCount),
		versionHasher: sha256.New(),
	}
}

func (r *resolveResponse) useResourceVersion(v string) error {
	_, err := r.versionHasher.Write([]byte(v))

	return err
}

func (r *resolveResponse) versionInfo() string {
	return base64.StdEncoding.EncodeToString(r.versionHasher.Sum(nil))
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
