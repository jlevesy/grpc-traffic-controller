package kxds

import (
	"context"
	"sync"
)

// resourceRef is a reference to an xDS resource.
type resourceRef struct {
	typeURL      string
	resourceName string
}

// watchBuilder allows to build a single watch, it returns the watch and a cleanup function.
type watchBuilder interface {
	buildWatch() (*watcher, func())
}

// watcher allows to subscribe and receive updates on subscribed resources.
type watcher struct {
	changes          chan resourceRef
	done             chan struct{}
	watchedResources map[resourceRef]struct{}

	watches *watches
}

func (w *watcher) notifyChanged(ctx context.Context, ref resourceRef) {
	select {
	case <-ctx.Done():
		return
	case <-w.done:
		return
	case w.changes <- ref:
	}
}

func (w *watcher) watch(ref resourceRef) {
	w.watches.getOrCreateResourceWatchers(ref).watch(w)
	w.watchedResources[ref] = struct{}{}
}

// resourceWatchers groups all watchers subscribing to a specific resource.
type resourceWatchers struct {
	mu       sync.RWMutex
	watchers map[*watcher]struct{}
}

func (rw *resourceWatchers) notifyChanged(ctx context.Context, ref resourceRef) {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	for w := range rw.watchers {
		go w.notifyChanged(ctx, ref)
	}
}

func (rw *resourceWatchers) watch(w *watcher) {
	rw.mu.RLock()
	_, ok := rw.watchers[w]
	rw.mu.RUnlock()

	if ok {
		return
	}

	rw.mu.Lock()
	rw.watchers[w] = struct{}{}
	rw.mu.Unlock()
}

func (rw *resourceWatchers) stopWatch(w *watcher) {
	rw.mu.RLock()
	_, ok := rw.watchers[w]
	rw.mu.RUnlock()

	if !ok {
		return
	}

	rw.mu.Lock()
	delete(rw.watchers, w)
	rw.mu.Unlock()
}

// watches keep track of all wachers grouped by resourceRef and allows to delivers notification to them.
type watches struct {
	mu       sync.RWMutex
	watchers map[resourceRef]*resourceWatchers
}

func newWatches() *watches {
	return &watches{watchers: make(map[resourceRef]*resourceWatchers)}
}

func (w *watches) buildWatch() (*watcher, func()) {
	newWatcher := &watcher{
		changes:          make(chan resourceRef),
		done:             make(chan struct{}),
		watchedResources: make(map[resourceRef]struct{}),
		watches:          w,
	}

	return newWatcher, func() {
		w.stopAllWatches(newWatcher)
		close(newWatcher.done)
	}
}

func (w *watches) getResourceWatchers(ref resourceRef) (*resourceWatchers, bool) {
	w.mu.RLock()
	rw, ok := w.watchers[ref]
	w.mu.RUnlock()

	return rw, ok
}

func (w *watches) getOrCreateResourceWatchers(ref resourceRef) *resourceWatchers {
	rw, ok := w.getResourceWatchers(ref)

	if ok {
		return rw
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	rw, ok = w.watchers[ref]
	if ok {
		return rw
	}

	rw = &resourceWatchers{watchers: make(map[*watcher]struct{})}

	w.watchers[ref] = rw

	return rw
}

func (w *watches) stopAllWatches(wa *watcher) {
	for ref := range wa.watchedResources {
		rw, ok := w.getResourceWatchers(ref)
		if !ok {
			continue
		}

		rw.stopWatch(wa)
	}
}

func (w *watches) notifyChanged(ctx context.Context, ref resourceRef) {
	rw, ok := w.getResourceWatchers(ref)
	if !ok {
		return
	}

	rw.notifyChanged(ctx, ref)
}
