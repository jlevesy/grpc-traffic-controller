package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	_ "google.golang.org/grpc/xds"

	echo "github.com/jlevesy/kxds/pkg/echoserver/proto"
)

type metadataArgs map[string]string

func (h metadataArgs) String() string {
	var sb strings.Builder

	for k, v := range h {
		sb.WriteString(k)
		sb.WriteRune('=')
		sb.WriteString(v)
	}

	return sb.String()
}

func (h metadataArgs) Set(v string) error {
	sp := strings.Split(v, "=")
	if len(sp) != 2 {
		return fmt.Errorf("malformed argument %q", v)
	}

	h[sp[0]] = sp[1]

	return nil
}

func main() {
	var (
		ctx  = context.Background()
		meta = make(metadataArgs)

		addr    string
		period  time.Duration
		premium bool
	)

	flag.StringVar(&addr, "addr", "localhost:3333", "the address to connect to")
	flag.DurationVar(&period, "period", 0*time.Second, "period to make calls")
	flag.BoolVar(&premium, "premium", false, "call premium")
	flag.Var(&meta, "metadata", "add metadata to the call")
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

	callCtx := metadata.NewOutgoingContext(
		ctx,
		metadata.New(meta),
	)

	callPolicy(func() {
		log.Println("Calling echo server")

		var (
			resp *echo.EchoReply
			err  error
		)

		if premium {
			resp, err = client.EchoPremium(callCtx, &echo.EchoRequest{Payload: flag.Arg(0)})
		} else {
			resp, err = client.Echo(callCtx, &echo.EchoRequest{Payload: flag.Arg(0)})
		}
		if err != nil {
			log.Println("unable to send echo request", err)
			return
		}

		log.Println("Received a response from:", resp.ServerId, ".Payload is:", resp.Payload, "Variant is:", resp.Variant)
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
