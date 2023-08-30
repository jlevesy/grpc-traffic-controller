package bootstrap

import (
	"context"
	"os"
)

type envProvider struct{}

func (e *envProvider) Provide(_ context.Context, serverURI string) (*BootstrapConfig, error) {
	nodeID, err := getNodeID()
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
				Zone: os.Getenv("GTC_ZONE"),
			},
		},
	}, nil
}

func getNodeID() (string, error) {
	nodeID := os.Getenv("GTC_NODE_ID")

	if nodeID != "" {
		return nodeID, nil
	}

	return os.Hostname()
}
