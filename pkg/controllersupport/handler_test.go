package controllersupport_test

import (
	"context"
	"testing"
	"time"

	"github.com/jlevesy/grpc-traffic-controller/pkg/controllersupport"
	"github.com/jlevesy/grpc-traffic-controller/pkg/testruntime"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestQueuedEventHandler_HandlesEvent(t *testing.T) {
	var (
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		handler     = testHandler{
			wantCalls: 3,
			done:      cancel,
		}
		eventHandler = controllersupport.NewQueuedEventHandler(
			&handler,
			1,
			"test",
			zap.NewNop(),
		)
	)

	defer cancel()

	// Enqueue some calls.
	eventHandler.OnAdd(testruntime.Ptr(4), true)
	// Make sure that invalid calls are ignored...
	eventHandler.OnUpdate(testruntime.Ptr(4), testruntime.Ptr(4))
	eventHandler.OnDelete(testruntime.Ptr(4))

	// Run the queue until completion.
	eventHandler.Run(ctx)

	assert.True(t, handler.isComplete())
}

type testHandler struct {
	addReceived    int
	updateReceived int
	deleteReceived int

	wantCalls int
	done      func()
}

func (h *testHandler) OnAdd(context.Context, any) error {
	h.addReceived++
	// On first call, return transient error to test the retry behavior.
	h.call()
	return nil
}

func (h *testHandler) OnUpdate(_ context.Context, _, _ any) error {
	h.updateReceived++
	h.call()
	return nil
}

func (h *testHandler) OnDelete(context.Context, any) error {
	h.deleteReceived++
	h.call()
	return nil
}

func (h *testHandler) call() {
	if h.isComplete() {
		h.done()
	}
}

func (h *testHandler) isComplete() bool {
	return h.wantCalls == (h.addReceived + h.updateReceived + h.deleteReceived)
}
