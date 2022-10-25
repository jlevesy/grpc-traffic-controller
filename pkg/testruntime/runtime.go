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
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type Runtime struct {
	Backends []Backend
	Client   *fake.ClientBuilder
}

type Config struct {
	BackendCount int
}

func New(cfg Config) (*Runtime, error) {
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

	return &Runtime{
		Backends: backends,
		Client:   fake.NewClientBuilder().WithScheme(scheme.Scheme),
	}, nil
}

func (r *Runtime) Stop() error {
	for _, b := range r.Backends {
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

func (b *Backend) Close() error {
	b.Server.Stop()
	return b.Listener.Close()
}
