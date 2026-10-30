package riidoaiserver

import "sync"

// Shared by quiet Store copies. Notifications carry no assignment payload and
// wake only subscribers for the affected agent. A full buffer coalesces hints.
type assignmentChangeHub struct {
	mu          sync.Mutex
	subscribers map[string]map[chan struct{}]struct{}
}

func newAssignmentChangeHub() *assignmentChangeHub {
	return &assignmentChangeHub{subscribers: map[string]map[chan struct{}]struct{}{}}
}

func (h *assignmentChangeHub) subscribe(ids []string) (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	h.mu.Lock()
	for _, id := range ids {
		if h.subscribers[id] == nil {
			h.subscribers[id] = map[chan struct{}]struct{}{}
		}
		h.subscribers[id][ch] = struct{}{}
	}
	h.mu.Unlock()
	var once sync.Once
	return ch, func() {
		once.Do(func() {
			h.mu.Lock()
			defer h.mu.Unlock()
			for _, id := range ids {
				delete(h.subscribers[id], ch)
				if len(h.subscribers[id]) == 0 {
					delete(h.subscribers, id)
				}
			}
		})
	}
}

func (h *assignmentChangeHub) signal(id string) {
	if h == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subscribers[id] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
