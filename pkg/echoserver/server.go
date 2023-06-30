package echoserver

import (
	"context"
	"sync"

	echo "github.com/jlevesy/grpc-traffic-controller/pkg/echoserver/proto"
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

type SwapableServer struct {
	echo.UnimplementedEchoServer

	mu   sync.RWMutex
	impl echo.EchoServer
}

func (s *SwapableServer) Swap(impl echo.EchoServer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.impl = impl
}

func (s *SwapableServer) Echo(ctx context.Context, req *echo.EchoRequest) (*echo.EchoReply, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.impl.Echo(ctx, req)
}

func (s *SwapableServer) EchoPremium(ctx context.Context, req *echo.EchoRequest) (*echo.EchoReply, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.impl.EchoPremium(ctx, req)
}
