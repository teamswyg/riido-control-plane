package riidoaiserver

func (s *Store) checkStateAssignmentEventACK(state *storeState, assignment Assignment, req AgentEventRequest) (AgentEventResponse, bool, error) {
	checker := *s
	if len(checker.eventACKFence) == 0 {
		checker.eventACKFence = state.assignmentEventACKs[assignment.ID]
	}
	return checker.checkAssignmentEventACK(assignment, req)
}

func (s *Store) retainStateAssignmentEventACK(state *storeState, current Assignment, req AgentEventRequest, event TaskEvent, payloadHash string) {
	s.eventACK = newAssignmentEventACKReceipt(current, req, event, payloadHash)
	if s.eventACK == nil {
		return
	}
	if state.assignmentEventACKs == nil {
		state.assignmentEventACKs = make(map[string][]assignmentEventACKReceipt)
	}
	state.assignmentEventACKs[current.ID] = append(state.assignmentEventACKs[current.ID], *s.eventACK)
}
