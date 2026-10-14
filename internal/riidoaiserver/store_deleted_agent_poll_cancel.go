package riidoaiserver

import "strings"

func deletedAgentPollCancelAssignment(state *storeState, agentID string) (Assignment, bool) {
	for _, assignmentID := range state.agentAssignments[agentID] {
		assignment := state.assignments[assignmentID]
		if assignment.State.Code() == AssignmentStateCodeCancelling {
			return assignment, true
		}
	}
	return Assignment{}, false
}

func deletedAgentPollCancelResponse(state *storeState, agentID string, req PollRequest, count bool) (PollResponse, bool) {
	if strings.TrimSpace(req.DaemonID) == "" || strings.TrimSpace(req.RuntimeID) == "" {
		return PollResponse{}, false
	}
	assignment, ok := deletedAgentPollCancelAssignment(state, agentID)
	if !ok {
		return PollResponse{}, false
	}
	response := PollResponse{
		SchemaVersion: SchemaVersion,
		Action:        PollCancel,
		Assignment:    copyAssignment(assignment),
	}
	recordPollAction(state, response.Action, count)
	return response, true
}
