package testruntime

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	echo "github.com/jlevesy/grpc-traffic-controller/pkg/echoserver/proto"
)

func MultiAssert(asserts ...func(t *testing.T, callCtx *CallContext)) func(t *testing.T, callCtx *CallContext) {
	return func(t *testing.T, callCtx *CallContext) {
		for _, assert := range asserts {
			assert(t, callCtx)
		}
	}
}

func Wait(d time.Duration) func(*testing.T, *CallContext) {
	return func(*testing.T, *CallContext) {
		time.Sleep(d)
	}
}

func ExceedDelay(d time.Duration, assert func(*testing.T, *CallContext)) func(*testing.T, *CallContext) {
	return func(t *testing.T, callCtx *CallContext) {
		start := time.Now()
		assert(t, callCtx)

		if time.Since(start) <= d {
			t.Fatal("Test duration did not exceed wanted duration")
		}
	}
}

func WithinDelay(d time.Duration, assert func(*testing.T, *CallContext)) func(*testing.T, *CallContext) {
	return func(t *testing.T, callCtx *CallContext) {
		start := time.Now()

		assert(t, callCtx)

		if time.Since(start) > d {
			t.Fatal("Test duration did not happen within wanted duration")
		}
	}
}

type Caller struct {
	m       Method
	req     *echo.EchoRequest
	ctx     context.Context
	timeout time.Duration
}

func (c *Caller) Do(cl echo.EchoClient) (*echo.EchoReply, error) {
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()

	return c.m(ctx, cl, c.req)
}

type CallerOpt func(c *Caller)

func WithTimeout(d time.Duration) CallerOpt {
	return func(c *Caller) {
		c.timeout = d
	}
}

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
		ctx:     context.Background(),
		timeout: 10 * time.Second,
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

type CallContext struct {
	addr   string
	conn   *grpc.ClientConn
	client echo.EchoClient
}

func (c *CallContext) Close() error {
	return c.conn.Close()
}

func DefaultCallContext(addr string) func(t *testing.T) *CallContext {
	return func(t *testing.T) *CallContext {
		conn, err := grpc.Dial(
			addr,
			grpc.WithTransportCredentials(
				insecure.NewCredentials(),
			),
		)
		require.NoError(t, err)

		return &CallContext{
			addr:   addr,
			conn:   conn,
			client: echo.NewEchoClient(conn),
		}
	}
}

func CallOnce(caller Caller, assertions ...CallsAssertion) func(t *testing.T, callCtx *CallContext) {
	return CallN(caller, 1, assertions...)
}

func CallN(caller Caller, count int, assertions ...CallsAssertion) func(t *testing.T, callCtx *CallContext) {
	return func(t *testing.T, callCtx *CallContext) {
		calls := make([]call, count)

		for i := 0; i < count; i++ {
			var c call

			resp, err := caller.Do(callCtx.client)

			c.addr = callCtx.addr
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

func CallNParallel(caller Caller, count int, assertions ...CallsAssertion) func(t *testing.T, callCtx *CallContext) {
	return func(t *testing.T, callCtx *CallContext) {
		var (
			calls = make([]call, count)

			group errgroup.Group
		)

		for i := 0; i < count; i++ {
			i := i
			group.Go(func() error {
				var (
					c call
				)

				resp, err := caller.Do(callCtx.client)

				c.addr = callCtx.addr
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

func CountByBackendID(asserts ...AggregatedCallAssertion) CallsAssertion {
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

func AssertCount(backendID string, wantCount int) AggregatedCallAssertion {
	return func(t *testing.T, aggs map[string]int) {
		assert.Equal(t, wantCount, aggs[backendID], backendID)
	}
}

func AssertAggregatedValuePartial(partial string, wantCount int) AggregatedCallAssertion {
	return func(t *testing.T, aggs map[string]int) {
		var matchCount int

		for k, v := range aggs {
			if strings.Contains(k, partial) {
				matchCount += v
			}
		}

		assert.Equal(t, wantCount, matchCount, partial)
	}
}

func AssertCountWithinDelta(backendID string, wantCount int, delta float64) AggregatedCallAssertion {
	return func(t *testing.T, aggs map[string]int) {
		assert.InDelta(t, wantCount, aggs[backendID], delta, backendID)
	}
}
