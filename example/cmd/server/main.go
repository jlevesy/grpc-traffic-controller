package main

import (
	"context"
	"flag"
	"log"
	"net"

	"github.com/jlevesy/kxds/example/pkg/echo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	var bindAddress string

	flag.StringVar(&bindAddress, "bind-address", ":3333", "server bind address")

	flag.Parse()

	srv := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))

	listener, err := net.Listen("tcp", bindAddress)
	if err != nil {
		log.Fatal("Can't bind to addr", bindAddress)
	}

	echo.RegisterEchoServer(
		srv,
		&echoService{},
	)

	log.Println("Server listening on 0.0.0.0:3333")
	if err := srv.Serve(listener); err != nil {
		log.Fatal("Serve returned an error: ", err)
	}
}

type echoService struct {
	echo.UnimplementedEchoServer
}

func (e *echoService) Echo(ctx context.Context, req *echo.EchoRequest) (*echo.EchoReply, error) {
	log.Println("Received a request", req.Payload)

	return &echo.EchoReply{Payload: req.Payload}, nil
}
