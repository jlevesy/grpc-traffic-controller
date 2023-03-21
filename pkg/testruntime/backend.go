package testruntime

import (
	"net"
	"strconv"
	"time"

	"github.com/jlevesy/kxds/pkg/echoserver"
	echo "github.com/jlevesy/kxds/pkg/echoserver/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Backends []Backend

type Config struct {
	BackendCount int
}

func StartBackends(cfg Config) (Backends, error) {
	var (
		err      error
		backends = make([]Backend, cfg.BackendCount)
	)

	for id := 0; id < cfg.BackendCount; id++ {
		backends[id], err = newBackend("backend-" + strconv.Itoa(id))
		if err != nil {
			return nil, err
		}
	}

	return Backends(backends), nil
}

func (bs Backends) Stop() error {
	for _, b := range bs {
		_ = b.Close()
	}

	return nil
}

func (bs Backends) SetBehavior(bh Behavior) {
	for _, b := range bs {
		b.SetBehavior(bh)
	}
}

type Behavior func(string) echo.EchoServer

func DefaultBehavior() Behavior {
	return func(id string) echo.EchoServer {
		return &echoserver.Server{
			EchoFunc: func(req *echo.EchoRequest) (*echo.EchoReply, error) {
				return &echo.EchoReply{ServerId: id, Payload: req.Payload, Variant: "standard"}, nil
			},
			EchoPremiumFunc: func(req *echo.EchoRequest) (*echo.EchoReply, error) {
				return &echo.EchoReply{ServerId: id, Payload: req.Payload, Variant: "premium"}, nil
			},
		}
	}
}

func HangBehavior(d time.Duration) Behavior {
	return func(id string) echo.EchoServer {
		return &echoserver.Server{
			EchoFunc: func(req *echo.EchoRequest) (*echo.EchoReply, error) {
				time.Sleep(d)
				return &echo.EchoReply{ServerId: id, Payload: req.Payload, Variant: "standard"}, nil
			},
			EchoPremiumFunc: func(req *echo.EchoRequest) (*echo.EchoReply, error) {
				time.Sleep(d)
				return &echo.EchoReply{ServerId: id, Payload: req.Payload, Variant: "premium"}, nil
			},
		}
	}
}

type Backend struct {
	ID       string
	Listener net.Listener
	Server   *grpc.Server
	Impl     *echoserver.SwapableServer
}

func newBackend(id string) (Backend, error) {
	srv := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return Backend{}, err
	}

	b := Backend{
		ID:       id,
		Listener: listener,
		Server:   srv,
		Impl:     &echoserver.SwapableServer{},
	}

	echo.RegisterEchoServer(srv, b.Impl)

	b.SetBehavior(DefaultBehavior())

	go func() {
		_ = srv.Serve(listener)
	}()

	return b, nil
}

func (b *Backend) SetBehavior(bh Behavior) { b.Impl.Swap(bh(b.ID)) }

func (b *Backend) PortNumber() int32 {
	_, p, err := net.SplitHostPort(b.Listener.Addr().String())
	if err != nil {
		panic(err)
	}

	pint, err := strconv.ParseInt(p, 10, 32)
	if err != nil {
		panic(err)
	}

	return int32(pint)
}
func (b *Backend) Close() error {
	b.Server.Stop()
	return b.Listener.Close()
}

type NoopCacheLogger struct{}

func (NoopCacheLogger) Debugf(format string, args ...interface{}) {}
func (NoopCacheLogger) Infof(format string, args ...interface{})  {}
func (NoopCacheLogger) Warnf(format string, args ...interface{})  {}
func (NoopCacheLogger) Errorf(format string, args ...interface{}) {}
