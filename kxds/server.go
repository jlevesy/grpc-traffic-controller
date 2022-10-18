package kxds

import (
	"context"
	"fmt"
	"net"
	"time"

	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/test/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	grpcKeepaliveTime        = 30 * time.Second
	grpcKeepaliveTimeout     = 5 * time.Second
	grpcKeepaliveMinTime     = 30 * time.Second
	grpcMaxConcurrentStreams = 1000000
)

type XDSServerConfig struct {
	BindAddr string
}

type XDSServer struct {
	xdsCache cache.Cache
	cfg      XDSServerConfig
}

func NewXDSServer(cache cache.Cache, cfg XDSServerConfig) *XDSServer {
	return &XDSServer{
		xdsCache: cache,
		cfg:      cfg,
	}
}

func (s *XDSServer) Start(ctx context.Context) error {
	var (
		logger = log.FromContext(ctx)
		cb     = test.Callbacks{Debug: true}
		server = server.NewServer(ctx, s.xdsCache, &cb)
	)

	grpcServer := grpc.NewServer(
		grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    grpcKeepaliveTime,
			Timeout: grpcKeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             grpcKeepaliveMinTime,
			PermitWithoutStream: true,
		}),
	)

	discoveryv3.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)

	logger.Info("Starting xDS server", "bindAddress", s.cfg.BindAddr)

	lis, err := net.Listen("tcp", s.cfg.BindAddr)
	if err != nil {
		return fmt.Errorf("unable to bind %w", err)
	}

	defer lis.Close()

	go func() {
		<-ctx.Done()

		logger.Info("Manager signaled termination, stopping the server")

		grpcServer.GracefulStop()

		logger.Info("server stopped")
	}()

	logger.Info("xDS server ready")

	return grpcServer.Serve(lis)
}
