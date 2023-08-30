package bootstrap

import (
	"context"
	"fmt"
	"strings"
)

const (
	ProviderTypeEnv    = "env"
	ProviderTypeGcloud = "gcloud"
)

type UnknownProviderError string

func (u UnknownProviderError) Error() string {
	return fmt.Sprintf("unknown bootstrap provider %q", string(u))
}

// ConfigProvider allows to retrieve a BootstrapConfig.
type ConfigProvider interface {
	Provide(ctx context.Context, serverURI string) (*BootstrapConfig, error)
}

func BuildConfigProvider(_ context.Context, providerType string) (ConfigProvider, error) {
	switch strings.ToLower(providerType) {
	case ProviderTypeEnv:
		return &envProvider{}, nil
	case ProviderTypeGcloud:
		return &gcloudProvider{}, nil
	default:
		return nil, UnknownProviderError(providerType)
	}
}
