package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/test/v3"
	pkglog "github.com/jlevesy/kxds/pkg/log"
	"github.com/jlevesy/kxds/pkg/routing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	grpcKeepaliveTime        = 30 * time.Second
	grpcKeepaliveTimeout     = 5 * time.Second
	grpcKeepaliveMinTime     = 30 * time.Second
	grpcMaxConcurrentStreams = 1000000
)

const kxdsHashKey = "kxds"

type constantHash string

func (h constantHash) ID(*core.Node) string { return string(h) }

func main() {
	var (
		listenPort uint
		logger     pkglog.Logger
	)

	flag.UintVar(&listenPort, "port", 18000, "xDS management server port")
	flag.BoolVar(&logger.Debug, "debug", true, "log all the things")

	flag.Parse()

	// Create a cache
	cache := cache.NewSnapshotCache(true, constantHash(kxdsHashKey), &logger)

	// Create the snapshot that we'll serve to Envoy
	snapshot := routing.GenerateSnapshot()

	logger.Infof("will serve snapshot %+v", snapshot)

	// Add the snapshot to the cache
	if err := cache.SetSnapshot(context.Background(), kxdsHashKey, snapshot); err != nil {
		logger.Errorf("snapshot error %q for %+v", err, snapshot)
		os.Exit(1)
	}

	// Run the xDS server
	var (
		ctx    = context.Background()
		cb     = test.Callbacks{Debug: logger.Debug}
		server = server.NewServer(ctx, cache, &cb)
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

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", listenPort))
	if err != nil {
		log.Fatal(err)
	}

	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)

	logger.Infof("Management server listening on %d\n", listenPort)

	if err = grpcServer.Serve(lis); err != nil {
		log.Println(err)
	}
}
