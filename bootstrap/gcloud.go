package bootstrap

import (
	"context"
	"errors"

	"cloud.google.com/go/compute/metadata"
)

var ErrNotRunningOnGCE = errors.New("not running on GCE")

type gcloudProvider struct{}

func (e *gcloudProvider) Provide(_ context.Context, serverURI string) (*BootstrapConfig, error) {
	if !metadata.OnGCE() {
		return nil, ErrNotRunningOnGCE
	}

	nodeID, err := getNodeID()
	if err != nil {
		return nil, err
	}

	zone, err := metadata.Zone()
	if err != nil {
		return nil, err
	}

	return &BootstrapConfig{
		XDSServers: []XDSServer{
			{
				URI:      serverURI,
				Features: []string{"xds_v3"},
				Creds:    []Cred{{Type: "insecure"}},
			},
		},
		Node: Node{
			ID: nodeID,
			Locality: Locality{
				Zone: zone,
			},
		},
	}, nil
}
