package main

import (
	"flag"
	"log"
	"net"
	"os"

	"github.com/jlevesy/grpc-traffic-controller/pkg/echoserver"
	echo "github.com/jlevesy/grpc-traffic-controller/pkg/echoserver/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	var (
		bindAddress string
	)

	flag.StringVar(&bindAddress, "bind-address", ":3333", "server bind address")

	flag.Parse()

	hostName, err := os.Hostname()
	if err != nil {
		log.Fatal("Could not read hostname", hostName)
	}

	srv := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))

	listener, err := net.Listen("tcp", bindAddress)
	if err != nil {
		log.Fatal("Can't bind to addr", bindAddress)
	}

	echo.RegisterEchoServer(
		srv,
		&echoserver.Server{
			EchoFunc: func(req *echo.EchoRequest) (*echo.EchoReply, error) {
				log.Println("Received a request", req.Payload)

				return &echo.EchoReply{
					Payload:  req.Payload,
					ServerId: hostName,
					Variant:  "standard",
				}, nil
			},
			EchoPremiumFunc: func(req *echo.EchoRequest) (*echo.EchoReply, error) {
				log.Println("Received a request v2", req.Payload)

				return &echo.EchoReply{
					Payload:  req.Payload,
					ServerId: hostName,
					Variant:  "premium",
				}, nil
			},
		},
	)

	log.Println("Server listening on 0.0.0.0:3333")
	if err := srv.Serve(listener); err != nil {
		log.Fatal("Serve returned an error: ", err)
	}
}
