package testruntime

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	echo "github.com/jlevesy/kxds/pkg/echoserver/proto"
)

func MultiAssert(asserts ...func(t *testing.T)) func(t *testing.T) {
	return func(t *testing.T) {
		for _, assert := range asserts {
			assert(t)
		}
	}
}

type call struct {
	addr      string
	backendID string
	err       error
}

type Method func(cl echo.EchoClient, req *echo.EchoRequest) (*echo.EchoReply, error)

func MethodEcho(cl echo.EchoClient, req *echo.EchoRequest) (*echo.EchoReply, error) {
	return cl.Echo(context.Background(), req)
}

func MethodEchoPremium(cl echo.EchoClient, req *echo.EchoRequest) (*echo.EchoReply, error) {
	return cl.EchoPremium(context.Background(), req)
}

type CallsAssertion func(t *testing.T, calls []call)

func CallOnce(addr string, method Method, assertions ...CallsAssertion) func(t *testing.T) {
	return CallN(addr, method, 1, assertions...)
}

func CallN(addr string, method Method, count int, assertions ...CallsAssertion) func(t *testing.T) {
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

			resp, err := method(
				client,
				&echo.EchoRequest{Payload: "Hello there"},
			)

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

func NoCallErrors(t *testing.T, calls []call) {
	for _, c := range calls {
		require.NoError(t, c.err)
	}
}

type AggregatedCallAssertion func(t *testing.T, counts map[string]int)

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

func BackendCalledExact(backendID string, wantCount int) AggregatedCallAssertion {
	return func(t *testing.T, aggs map[string]int) {
		assert.Equal(t, wantCount, aggs[backendID])
	}
}

func BackendCalledDelta(backendID string, wantCount int, delta float64) AggregatedCallAssertion {
	return func(t *testing.T, aggs map[string]int) {
		assert.InDelta(t, wantCount, aggs[backendID], delta)
	}
}
