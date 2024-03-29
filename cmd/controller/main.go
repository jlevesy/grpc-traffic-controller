/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gtcinformers "github.com/jlevesy/grpc-traffic-controller/client/informers/externalversions"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	restclient "k8s.io/client-go/rest"

	gtcapi "github.com/jlevesy/grpc-traffic-controller/client/clientset/versioned"
	"github.com/jlevesy/grpc-traffic-controller/gtc"
)

func main() {
	var (
		xdsAddr  string
		httpAddr string
		logLevel string
	)

	flag.StringVar(&xdsAddr, "xds-bind-address", ":18000", "The address the xds server binds to.")
	flag.StringVar(&httpAddr, "http-bind-address", ":8081", "The address the http server binds to.")
	flag.StringVar(&logLevel, "log-level", "info", "Log Level")
	flag.Parse()

	logger := zap.Must(newLogger(logLevel))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	kubeConfig, err := restclient.InClusterConfig()
	if err != nil {
		logger.Error("Can't build kube config", zap.Error(err))
		return
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		logger.Error("Can't create kubernetes client", zap.Error(err))
		return
	}

	gtcClient, err := gtcapi.NewForConfig(kubeConfig)
	if err != nil {
		logger.Error("Can't create gtc client", zap.Error(err))
		return
	}

	var (
		kubeInformerFactory = kubeinformers.NewSharedInformerFactory(
			kubeClient,
			60*time.Minute,
		)
		gtcInformerFactory = gtcinformers.NewSharedInformerFactory(
			gtcClient,
			60*time.Minute,
		)
	)

	server, err := gtc.NewXDSServer(
		ctx,
		gtc.XDSServerConfig{
			BindAddr:     xdsAddr,
			K8sInformers: kubeInformerFactory,
			GTCInformers: gtcInformerFactory,
		},
		logger,
	)
	if err != nil {
		logger.Error("Can't create gtc server", zap.Error(err))
		return
	}

	group, ctx := errgroup.WithContext(ctx)

	logger.Info("Starting informers...")

	gtcInformerFactory.Start(ctx.Done())
	kubeInformerFactory.Start(ctx.Done())

	group.Go(func() error {
		return server.Run(ctx)
	})

	group.Go(func() error {
		return runWebserver(
			ctx,
			httpAddr,
			logger.With(
				zap.String("module", "webserver"),
			),
		)
	})

	logger.Info("Running gRPC Traffic controller")

	if err := group.Wait(); err != nil {
		logger.Error("Controller reported an error", zap.Error(err))
		return
	}

	logger.Info("gRPC Traffic controller exited")
}

func runWebserver(ctx context.Context, addr string, logger *zap.Logger) error {
	logger.Info("Starting webserver", zap.String("addr", addr))

	serveMux := http.NewServeMux()

	serveMux.HandleFunc("/healthz", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("ok"))
	})

	serveMux.HandleFunc("/readyz", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:           addr,
		Handler:        serveMux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1048576
	}

	go func() {
		<-ctx.Done()

		logger.Info("Shutting down webserver")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("Shutdown reported an error, closing the server", zap.Error(err))

			_ = srv.Close()
		}
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func newLogger(lvl string) (*zap.Logger, error) {
	if lvl == "debug" {
		return zap.NewDevelopment()
	}

	return zap.NewProduction()
}
