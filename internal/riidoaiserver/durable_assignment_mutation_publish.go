package riidoaiserver

func (s *Store) publishDurableAssignmentMutation(state, trial *storeState, assignment Assignment, events []TaskEvent, observe ...func(TaskEvent)) {
	applyAssignmentToState(state, assignment)
	state.agentEventsTotal = trial.agentEventsTotal
	for _, event := range events {
		s.appendRecordedEvent(state, event)
		for _, observer := range observe {
			observer(event)
		}
	}
	s.signalAgentWaiters(state, assignment.AgentID)
}
