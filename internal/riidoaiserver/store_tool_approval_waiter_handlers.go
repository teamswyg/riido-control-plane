package riidoaiserver

func (s *Store) handleRegisterToolApprovalWaiter(state *storeState, key string) (chan struct{}, int64) {
	state.nextToolApprovalWaiter++
	id := state.nextToolApprovalWaiter
	ch := make(chan struct{}, 1)
	if state.toolApprovalWaiters[key] == nil {
		state.toolApprovalWaiters[key] = map[int64]chan struct{}{}
	}
	state.toolApprovalWaiters[key][id] = ch
	return ch, id
}

func (s *Store) handleUnregisterToolApprovalWaiter(state *storeState, key string, id int64) {
	waiters := state.toolApprovalWaiters[key]
	if waiters == nil {
		return
	}
	delete(waiters, id)
	if len(waiters) == 0 {
		delete(state.toolApprovalWaiters, key)
	}
}

func (s *Store) signalToolApprovalWaiters(state *storeState, key string) {
	for _, ch := range state.toolApprovalWaiters[key] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
