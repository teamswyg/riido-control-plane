package riidoaiserver

// handleRegisterWaiter / handleUnregisterWaiter / signalAgentWaiters run only on
// the actor goroutine, so all access to state.agentWaiters is serialized and
// signal sends never race a register/unregister. The signal channel is buffered
// (cap 1) so a wake between two waiter selects is never lost; we never close it
// (the waiting goroutine only receives), so a stray late signal cannot panic.
func (s *Store) handleRegisterWaiter(state *storeState, agentID string) (chan struct{}, int64) {
	state.nextAgentWaiterSeq++
	id := state.nextAgentWaiterSeq
	ch := make(chan struct{}, 1)
	if state.agentWaiters[agentID] == nil {
		state.agentWaiters[agentID] = map[int64]chan struct{}{}
	}
	state.agentWaiters[agentID][id] = ch
	return ch, id
}

func (s *Store) handleUnregisterWaiter(state *storeState, agentID string, id int64) {
	waiters := state.agentWaiters[agentID]
	if waiters == nil {
		return
	}
	delete(waiters, id)
	if len(waiters) == 0 {
		delete(state.agentWaiters, agentID)
	}
}

// signalAgentWaiters wakes every long-poll parked on agentID. It is a one-shot,
// non-blocking broadcast: a waiter whose buffer is already full is left as-is
// (it will re-evaluate anyway). Called from the queued-producing transitions.
func (s *Store) signalAgentWaiters(state *storeState, agentID string) {
	for _, ch := range state.agentWaiters[agentID] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
