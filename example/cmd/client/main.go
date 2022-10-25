package main

import (
	"context"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/xds"

	echo "github.com/jlevesy/kxds/pkg/echoserver/proto"
)

func main() {
	var (
		ctx    = context.Background()
		addr   string
		period time.Duration
	)

	flag.StringVar(&addr, "addr", "localhost:3333", "the address to connect to")
	flag.DurationVar(&period, "period", 0*time.Second, "period to make calls")
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatal("Must provide a message")
	}

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}

	client := echo.NewEchoClient(conn)
	callPolicy := oneShot

	if period != 0 {
		callPolicy = repeated(period)
	}

	callPolicy(func() {
		log.Println("Calling echo server")

		resp, err := client.Echo(ctx, &echo.EchoRequest{Payload: flag.Arg(0)})
		if err != nil {
			log.Println("unable to send echo request", err)
			return
		}

		log.Println("Received echo response", resp.Payload)
	})
}

func oneShot(callback func()) { callback() }

func repeated(period time.Duration) func(func()) {
	return func(callback func()) {
		ticker := time.NewTicker(period)
		defer ticker.Stop()

		for {
			<-ticker.C
			callback()
		}
	}
}
