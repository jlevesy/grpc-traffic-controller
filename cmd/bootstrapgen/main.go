package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jlevesy/grpc-traffic-controller/bootstrap"
)

func main() {
	var (
		out       string
		serverURI string
		provider  string
	)

	flag.StringVar(&serverURI, "server-uri", "", "uri of the xds server")
	flag.StringVar(&out, "out", "./bootstrap.json", "path to write the generated config")
	flag.StringVar(&provider, "provider", bootstrap.ProviderTypeEnv, "provider to use")
	flag.Parse()

	if serverURI == "" {
		log.Fatal("please provide a server-uri")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	configProvider, err := bootstrap.BuildConfigProvider(ctx, provider)
	if err != nil {
		log.Fatal("unable to build config provider", err)
	}

	cfg, err := configProvider.Provide(ctx, serverURI)
	if err != nil {
		log.Fatal("unable to retrieve bootstrap config", err)
	}

	output, err := os.Create(out)
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()

	if err := json.NewEncoder(output).Encode(&cfg); err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully wrote configuration at path:", out)
}
