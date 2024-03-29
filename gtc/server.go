package gtc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	sotwv3 "github.com/envoyproxy/go-control-plane/pkg/server/sotw/v3"
	gtcinformers "github.com/jlevesy/grpc-traffic-controller/client/informers/externalversions"
	"github.com/jlevesy/grpc-traffic-controller/pkg/controllersupport"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	kubeinformers "k8s.io/client-go/informers"
)

const (
	grpcKeepaliveTime        = 30 * time.Second
	grpcKeepaliveTimeout     = 5 * time.Second
	grpcKeepaliveMinTime     = 30 * time.Second
	grpcMaxConcurrentStreams = 1000000
)

type XDSServerConfig struct {
	BindAddr     string
	K8sInformers kubeinformers.SharedInformerFactory
	GTCInformers gtcinformers.SharedInformerFactory
}

type XDSServer struct {
	bindAddr string
	server   *grpc.Server
	logger   *zap.Logger

	grpcListenerChangedQueue  *controllersupport.QueuedEventHandler
	endpointSliceChangedQueue *controllersupport.QueuedEventHandler
}

func NewXDSServer(ctx context.Context, cfg XDSServerConfig, logger *zap.Logger) (*XDSServer, error) {
	var (
		grpcServer = grpc.NewServer(
			grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
			grpc.KeepaliveParams(
				keepalive.ServerParameters{
					Time:    grpcKeepaliveTime,
					Timeout: grpcKeepaliveTimeout,
				},
			),
			grpc.KeepaliveEnforcementPolicy(
				keepalive.EnforcementPolicy{
					MinTime:             grpcKeepaliveMinTime,
					PermitWithoutStream: true,
				},
			),
		)
		watches = newWatches()
		srv     = sotwv3.NewServer(
			ctx,
			newConfigWatcher(
				cfg.K8sInformers.Discovery().V1().EndpointSlices().Lister(),
				cfg.GTCInformers.Api().V1alpha1().GRPCListeners().Lister(),
				watches,
				logger,
			),
			&loggerCallbacks{l: logger},
		)

		grpcListenerChangedQueue = controllersupport.NewQueuedEventHandler(
			&grpcListenerChangedHandler{
				watches: watches,
				logger:  logger,
			},
			10,
			"grpc-listeners-changes",
			logger,
		)

		endpointSliceChangedQueue = controllersupport.NewQueuedEventHandler(
			&endpointSliceChangedHandler{
				listenersLister: cfg.GTCInformers.Api().V1alpha1().GRPCListeners().Lister(),
				watches:         watches,
				logger:          logger,
			},
			10,
			"endpointslices-changes",
			logger,
		)
	)

	discoveryv3.RegisterAggregatedDiscoveryServiceServer(
		grpcServer, &adsHandler{srv: srv},
	)

	_, err := cfg.GTCInformers.
		Api().
		V1alpha1().
		GRPCListeners().
		Informer().
		AddEventHandler(grpcListenerChangedQueue)
	if err != nil {
		return nil, err
	}

	_, err = cfg.K8sInformers.
		Discovery().
		V1().
		EndpointSlices().
		Informer().
		AddEventHandler(endpointSliceChangedQueue)
	if err != nil {
		return nil, err
	}

	return &XDSServer{
		grpcListenerChangedQueue:  grpcListenerChangedQueue,
		endpointSliceChangedQueue: endpointSliceChangedQueue,
		bindAddr:                  cfg.BindAddr,
		server:                    grpcServer,
		logger:                    logger,
	}, nil
}

func (s *XDSServer) Run(ctx context.Context) error {
	errGroup, groupCtx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		s.grpcListenerChangedQueue.Run(groupCtx)
		return nil
	})

	errGroup.Go(func() error {
		s.endpointSliceChangedQueue.Run(groupCtx)
		return nil
	})

	errGroup.Go(func() error {
		lis, err := net.Listen("tcp", s.bindAddr)
		if err != nil {
			return fmt.Errorf("unable to bind %w", err)
		}

		defer lis.Close()

		go func() {
			<-groupCtx.Done()

			s.server.GracefulStop()

			s.logger.Info("gRPC server has stopped")
		}()

		return s.server.Serve(lis)
	})

	return errGroup.Wait()
}

type adsHandler struct {
	srv sotwv3.Server
}

func (h *adsHandler) StreamAggregatedResources(stream discoveryv3.AggregatedDiscoveryService_StreamAggregatedResourcesServer) error {
	return h.srv.StreamHandler(stream, resource.AnyType)
}

func (h *adsHandler) DeltaAggregatedResources(discoveryv3.AggregatedDiscoveryService_DeltaAggregatedResourcesServer) error {
	return errors.New("unsupported")
}
