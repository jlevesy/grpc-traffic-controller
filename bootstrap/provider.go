package bootstrap

import (
	"context"
	"fmt"
	"strings"
)

const (
	ProviderTypeEnv = "env"
)

type UnknownProviderError string

func (u UnknownProviderError) Error() string {
	return fmt.Sprintf("unknown bootstrap provider %q", string(u))
}

// ConfigProvider allows to retrieve a BootstrapConfig.
type ConfigProvider interface {
	Provide(ctx context.Context, serverURI string) (*BootstrapConfig, error)
}

func BuildConfigProvider(ctx context.Context, providerType string) (ConfigProvider, error) {
	switch strings.ToLower(providerType) {
	case ProviderTypeEnv:
		return &envProvider{}, nil
	default:
		return nil, UnknownProviderError(providerType)
	}
}
