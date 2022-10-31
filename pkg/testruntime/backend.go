package testruntime

import (
	"net"
	"strconv"

	kxdsv1alpha1 "github.com/jlevesy/kxds/api/v1alpha1"
	"github.com/jlevesy/kxds/pkg/echoserver"
	echo "github.com/jlevesy/kxds/pkg/echoserver/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
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

	if err = corev1.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}

	if err = kxdsv1alpha1.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}

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

type Backend struct {
	ID       string
	Listener net.Listener
	Server   *grpc.Server
}

func newBackend(id string) (Backend, error) {
	srv := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return Backend{}, err
	}

	echo.RegisterEchoServer(
		srv,
		&echoserver.Server{
			EchoFunc: func(req *echo.EchoRequest) (*echo.EchoReply, error) {
				return &echo.EchoReply{ServerId: id, Payload: req.Payload}, nil
			},
		},
	)

	go func() {
		_ = srv.Serve(listener)
	}()

	return Backend{
		ID:       id,
		Listener: listener,
		Server:   srv,
	}, nil
}

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
