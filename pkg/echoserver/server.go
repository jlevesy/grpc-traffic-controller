package echoserver

import (
	"context"

	echo "github.com/jlevesy/kxds/pkg/echoserver/proto"
)

type Server struct {
	echo.UnimplementedEchoServer

	EchoFunc        func(*echo.EchoRequest) (*echo.EchoReply, error)
	EchoPremiumFunc func(*echo.EchoRequest) (*echo.EchoReply, error)
}

func (e *Server) Echo(ctx context.Context, req *echo.EchoRequest) (*echo.EchoReply, error) {
	return e.EchoFunc(req)
}

func (e *Server) EchoPremium(ctx context.Context, req *echo.EchoRequest) (*echo.EchoReply, error) {
	return e.EchoPremiumFunc(req)
}
