package controllersupport

import (
	"context"
	"time"

	"go.uber.org/zap"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
)

// EventHandler represents a object that handles mutation event  a k8s object.
// It is based on k8s-client cache.EventHandler.
// Each method returns an insght that tells how the worker should behave regarding that event.
type EventHandler interface {
	OnAdd(ctx context.Context, obj any) error
	OnUpdate(ctx context.Context, oldObj, newObj any) error
	OnDelete(ctx context.Context, obj any) error
}

// QueuedEventHandler implements cache.EventHandler over a workqueue.
// It also handles the type conversion of updated objects to pass them down to an EventHandler.
type QueuedEventHandler struct {
	workqueue workqueue.RateLimitingInterface
	workers   int

	handler EventHandler
	logger  *zap.Logger
}

// NewQueuedEventHandler returns a queued event handler
func NewQueuedEventHandler(handler EventHandler, workers int, name string, logger *zap.Logger) *QueuedEventHandler {
	return &QueuedEventHandler{
		workqueue: workqueue.NewRateLimitingQueueWithConfig(
			workqueue.DefaultControllerRateLimiter(),
			workqueue.RateLimitingQueueConfig{Name: name},
		),

		handler: handler,
		workers: workers,
		logger:  logger.With(zap.String("queue_name", name)),
	}
}

// OnAdd enqueues an add event.
func (h *QueuedEventHandler) OnAdd(obj any, _ bool) {
	h.workqueue.Add(queueEvent{kind: kindAdd, object: obj})
}

// OnUpdate enqueues an update event.
func (h *QueuedEventHandler) OnUpdate(oldObj, newObj any) {
	h.workqueue.Add(queueEvent{kind: kindUpdate, oldObj: oldObj, newObj: newObj})
}

// OnDelete enqueues an update event.
func (h *QueuedEventHandler) OnDelete(obj any) {
	h.workqueue.Add(queueEvent{kind: kindDelete, object: obj})
}

// Run starts workers and waits until completion.
func (h *QueuedEventHandler) Run(ctx context.Context) {
	defer utilruntime.HandleCrash()
	defer h.workqueue.ShutDown()

	h.logger.Info("Starting workers", zap.Int("worker_count", h.workers))

	for i := 0; i < h.workers; i++ {
		go wait.UntilWithContext(ctx, h.runWorker, time.Second)
	}

	h.logger.Info("Started workers")

	<-ctx.Done()

	h.logger.Info("Shutting down workers")
}

func (h *QueuedEventHandler) runWorker(ctx context.Context) {
	for h.processItem(ctx) {
	}
}

func (h *QueuedEventHandler) processItem(ctx context.Context) bool {
	obj, shutdown := h.workqueue.Get()

	if shutdown {
		return false
	}

	defer h.workqueue.Done(obj)
	// Let's be naive and say that we don't really need to retry anything.
	defer h.workqueue.Forget(obj)

	event, ok := obj.(queueEvent)

	if !ok {
		return true
	}

	var err error

	switch event.kind {
	case kindAdd:
		err = h.handler.OnAdd(ctx, event.object)
	case kindUpdate:
		err = h.handler.OnUpdate(ctx, event.oldObj, event.newObj)
	case kindDelete:
		err = h.handler.OnDelete(ctx, event.object)

	default:
		return true
	}

	if err != nil {
		h.logger.Error("Event handler reported an error", zap.Error(err))
	}

	return true
}

type queueEventKind uint8

const (
	// forces assignation of an explicit value when using this enumeration.
	kindUnknown queueEventKind = iota //nolint:deadcode,varcheck
	kindAdd
	kindUpdate
	kindDelete
)

type queueEvent struct {
	kind   queueEventKind
	object any
	oldObj any
	newObj any
}
