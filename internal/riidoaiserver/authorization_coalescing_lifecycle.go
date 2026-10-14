package riidoaiserver

func (a *CoalescingAuthorizer) joinOrStart(key string) (*coalescedAuthorizationCall, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if call, ok := a.calls[key]; ok {
		return call, false
	}
	call := &coalescedAuthorizationCall{done: make(chan struct{})}
	a.calls[key] = call
	return call, true
}

func (a *CoalescingAuthorizer) finish(key string, call *coalescedAuthorizationCall) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.calls, key)
	close(call.done)
}
