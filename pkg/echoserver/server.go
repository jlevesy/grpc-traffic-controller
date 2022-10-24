package echoserver

import (
	"context"

	echo "github.com/jlevesy/kxds/pkg/echoserver/proto"
)

type Server struct {
	echo.UnimplementedEchoServer

	EchoFunc func(*echo.EchoRequest) (*echo.EchoReply, error)
}

func (e *Server) Echo(ctx context.Context, req *echo.EchoRequest) (*echo.EchoReply, error) {
	return e.EchoFunc(req)
}
