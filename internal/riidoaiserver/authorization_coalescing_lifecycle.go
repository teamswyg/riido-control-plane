package riidoaiserver

import "context"

func (a *CoalescingAuthorizer) joinOrStart(ctx context.Context, key string) (*coalescedAuthorizationCall, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if call, ok := a.calls[key]; ok {
		call.waiters++
		return call, false
	}
	shared, cancel := context.WithCancel(context.WithoutCancel(ctx))
	call := &coalescedAuthorizationCall{done: make(chan struct{}), ctx: shared, cancel: cancel, waiters: 1}
	a.calls[key] = call
	return call, true
}

func (a *CoalescingAuthorizer) finish(key string, call *coalescedAuthorizationCall) {
	a.mu.Lock()
	call.finished = true
	if a.calls[key] == call {
		delete(a.calls, key)
	}
	close(call.done)
	a.mu.Unlock()
	call.cancel()
}

func (a *CoalescingAuthorizer) release(key string, call *coalescedAuthorizationCall) {
	a.mu.Lock()
	call.waiters--
	abandoned := call.waiters == 0 && !call.finished
	if abandoned && a.calls[key] == call {
		delete(a.calls, key)
	}
	a.mu.Unlock()
	if abandoned {
		call.cancel()
	}
}
