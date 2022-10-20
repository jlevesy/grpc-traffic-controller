package main

import (
	"context"
	"flag"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/xds"

	"github.com/jlevesy/kxds/example/pkg/echo"
)

func main() {
	var (
		ctx  = context.Background()
		addr string
	)

	flag.StringVar(&addr, "addr", "localhost:3333", "the address to connect to")
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

	resp, err := client.Echo(ctx, &echo.EchoRequest{Payload: flag.Arg(0)})
	if err != nil {
		log.Fatal("unable to send echo request", err)
	}

	log.Println("Received echo message", resp.Payload)
}
