package kxds

import (
	"context"
	"sync"
)

type resourceRef struct {
	typeURL      string
	resourceName string
}

type watcher struct {
	changes            chan (resourceRef)
	done               chan struct{}
	watchedResourcesMu sync.RWMutex
	watchedResources   map[resourceRef]struct{}
}

func (w *watcher) notify(ctx context.Context, ref resourceRef) {
	w.watchedResourcesMu.RLock()
	_, ok := w.watchedResources[ref]
	w.watchedResourcesMu.RUnlock()

	if !ok {
		return
	}

	select {
	case w.changes <- ref:
	case <-ctx.Done():
		return
	case <-w.done:
		return
	}
}

func (w *watcher) watch(ref resourceRef) {
	w.watchedResourcesMu.RLock()
	_, ok := w.watchedResources[ref]
	w.watchedResourcesMu.RUnlock()

	// We're already watching that thing. Early exit and don't grab a write lock.
	if ok {
		return
	}

	w.watchedResourcesMu.Lock()
	defer w.watchedResourcesMu.Unlock()

	if _, ok := w.watchedResources[ref]; ok {
		return
	}

	w.watchedResources[ref] = struct{}{}
}

type watchBuilder interface {
	buildWatch() (*watcher, func())
}

type watches struct {
	mu       sync.RWMutex
	watchers map[*watcher]struct{}
}

func newWatches() *watches {
	return &watches{watchers: make(map[*watcher]struct{})}
}

func (w *watches) buildWatch() (*watcher, func()) {
	watcher := watcher{
		changes:          make(chan resourceRef),
		done:             make(chan struct{}),
		watchedResources: make(map[resourceRef]struct{}),
	}

	w.mu.Lock()
	w.watchers[&watcher] = struct{}{}
	w.mu.Unlock()

	return &watcher, func() {
		w.mu.Lock()
		delete(w.watchers, &watcher)
		close(watcher.done)
		w.mu.Unlock()
	}
}

func (w *watches) notifyChanged(ctx context.Context, ref resourceRef) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for watcher := range w.watchers {
		watcher := watcher

		go watcher.notify(ctx, ref)
	}
}
