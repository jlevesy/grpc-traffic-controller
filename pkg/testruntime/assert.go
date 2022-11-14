package testruntime

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	echo "github.com/jlevesy/kxds/pkg/echoserver/proto"
)

func MultiAssert(asserts ...func(t *testing.T)) func(t *testing.T) {
	return func(t *testing.T) {
		for _, assert := range asserts {
			assert(t)
		}
	}
}

func WithinDeadline(d time.Duration, assert func(t *testing.T)) func(t *testing.T) {
	return func(t *testing.T) {
		tt := time.NewTimer(d)
		defer tt.Stop()

		go func() {
			<-tt.C
			t.Error("Test did not succeed within deadline")
		}()

		assert(t)
	}
}

type Caller struct {
	m   Method
	req *echo.EchoRequest
	ctx context.Context
}

func (c *Caller) Do(cl echo.EchoClient) (*echo.EchoReply, error) {
	return c.m(c.ctx, cl, c.req)
}

type CallerOpt func(c *Caller)

func WithMetadata(meta map[string]string) CallerOpt {
	return WithContext(
		metadata.NewOutgoingContext(
			context.Background(),
			metadata.New(meta),
		),
	)
}

func WithContext(ctx context.Context) CallerOpt {
	return func(c *Caller) {
		c.ctx = ctx
	}
}

func BuildCaller(method Method, opts ...CallerOpt) Caller {
	caller := Caller{
		m: method,
		req: &echo.EchoRequest{
			Payload: "Hello There!",
		},
		ctx: context.Background(),
	}

	for _, opt := range opts {
		opt(&caller)
	}

	return caller
}

type Method func(ctx context.Context, cl echo.EchoClient, req *echo.EchoRequest) (*echo.EchoReply, error)

func MethodEcho(ctx context.Context, cl echo.EchoClient, req *echo.EchoRequest) (*echo.EchoReply, error) {
	return cl.Echo(ctx, req)
}

func MethodEchoPremium(ctx context.Context, cl echo.EchoClient, req *echo.EchoRequest) (*echo.EchoReply, error) {
	return cl.EchoPremium(ctx, req)
}

type call struct {
	addr      string
	backendID string
	err       error
}

type CallsAssertion func(t *testing.T, calls []call)

func CallOnce(addr string, caller Caller, assertions ...CallsAssertion) func(t *testing.T) {
	return CallN(addr, caller, 1, assertions...)
}
func CallN(addr string, caller Caller, count int, assertions ...CallsAssertion) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := grpc.Dial(
			addr,
			grpc.WithTransportCredentials(
				insecure.NewCredentials(),
			),
		)
		require.NoError(t, err)

		defer conn.Close()

		var (
			client = echo.NewEchoClient(conn)
			calls  = make([]call, count)
		)

		for i := 0; i < count; i++ {
			var c call

			resp, err := caller.Do(client)

			c.addr = addr
			c.err = err
			if err == nil {
				c.backendID = resp.ServerId
			}

			calls[i] = c
		}

		for _, assert := range assertions {
			assert(t, calls)
		}
	}
}

func CallNParallel(addr string, caller Caller, count int, assertions ...CallsAssertion) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := grpc.Dial(
			addr,
			grpc.WithTransportCredentials(
				insecure.NewCredentials(),
			),
		)
		require.NoError(t, err)

		defer conn.Close()

		var (
			client = echo.NewEchoClient(conn)
			calls  = make([]call, count)

			group errgroup.Group
		)

		for i := 0; i < count; i++ {
			i := i
			group.Go(func() error {
				var (
					c call
				)

				resp, err := caller.Do(client)

				c.addr = addr
				c.err = err

				if err == nil {
					c.backendID = resp.ServerId
				}

				calls[i] = c

				return nil
			})
		}

		require.NoError(t, group.Wait())

		for _, assert := range assertions {
			assert(t, calls)
		}
	}
}

func NoCallErrors(t *testing.T, calls []call) {
	for _, c := range calls {
		require.NoError(t, c.err)
	}
}

func MustFail(t *testing.T, calls []call) {
	for _, c := range calls {
		require.Error(t, c.err)
	}
}

type AggregatedCallAssertion func(t *testing.T, counts map[string]int)

func AggregateByError(asserts ...AggregatedCallAssertion) CallsAssertion {
	return func(t *testing.T, calls []call) {
		agg := make(map[string]int)

		for _, c := range calls {
			key := "ok"

			if c.err != nil {
				key = c.err.Error()
			}

			agg[key] += 1
		}

		for _, assert := range asserts {
			assert(t, agg)
		}
	}
}

func AggregateByBackendID(asserts ...AggregatedCallAssertion) CallsAssertion {
	return func(t *testing.T, calls []call) {
		agg := make(map[string]int)

		for _, c := range calls {
			if c.err != nil {
				continue
			}

			agg[c.backendID] += 1
		}

		for _, assert := range asserts {
			assert(t, agg)
		}
	}
}

func DumpCounts(t *testing.T, aggs map[string]int) {
	for k, v := range aggs {
		t.Logf("%s => %d", k, v)
	}
}

func AssertAggregatedValue(backendID string, wantCount int) AggregatedCallAssertion {
	return func(t *testing.T, aggs map[string]int) {
		assert.Equal(t, wantCount, aggs[backendID], backendID)
	}
}

func AssertAggregatedValueWithinDelta(backendID string, wantCount int, delta float64) AggregatedCallAssertion {
	return func(t *testing.T, aggs map[string]int) {
		assert.InDelta(t, wantCount, aggs[backendID], delta, backendID)
	}
}
