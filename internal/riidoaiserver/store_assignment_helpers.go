package riidoaiserver

import "slices"

func assignmentHoldsActiveLease(state AssignmentState) bool {
	code := state.Code()
	return code.IsAgentActive() || code == AssignmentStateCodeCancelling
}

func assignmentBlockerCleared(state *storeState, assignment Assignment) bool {
	if assignment.BlockedByAssignmentID == "" {
		return true
	}
	blocker := state.assignments[assignment.BlockedByAssignmentID]
	return isTerminal(blocker.State)
}

func (state *storeState) assignmentForClientStop(taskID string, req CancelAssignmentRequest) (Assignment, bool) {
	if req.AssignmentID != "" {
		assignment := state.assignments[req.AssignmentID]
		return assignment, assignment.ID != ""
	}
	assignmentIDs := state.agentAssignments[req.AgentID]
	for _, assignmentID := range slices.Backward(assignmentIDs) {
		assignment := state.assignments[assignmentID]
		if assignment.TaskID != taskID || isTerminal(assignment.State) {
			continue
		}
		return assignment, true
	}
	return Assignment{}, false
}

func copyAssignment(a Assignment) *Assignment {
	cp := a
	return &cp
}

func assignmentIDInAgentQueue(ids []string, id string) bool {
	return slices.Contains(ids, id)
}
