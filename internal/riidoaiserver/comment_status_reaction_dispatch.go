package riidoaiserver

import "sync"

// commentStatusReactionDispatcher runs at most one writer per thread outside
// the store locks. A request that arrives while the writer runs makes it read
// the latest plan again, so the last written reaction follows the last state.
type commentStatusReactionDispatcher struct {
	mu      sync.Mutex
	running map[string]bool
	pending map[string]bool
	wg      sync.WaitGroup
}

func newCommentStatusReactionDispatcher() *commentStatusReactionDispatcher {
	return &commentStatusReactionDispatcher{running: map[string]bool{}, pending: map[string]bool{}}
}

func (d *commentStatusReactionDispatcher) start(key string, run func()) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.running[key] {
		d.pending[key] = true
		return
	}
	d.running[key] = true
	d.wg.Go(func() {
		for {
			run()
			d.mu.Lock()
			if d.pending[key] {
				delete(d.pending, key)
				d.mu.Unlock()
				continue
			}
			delete(d.running, key)
			d.mu.Unlock()
			return
		}
	})
}
