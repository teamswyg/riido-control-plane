package riidoaiserver

import "strings"

func heartbeatAssignmentIDs(state *storeState, agentID string, req AgentHeartbeatRequest) []string {
	seen := map[string]bool{}
	var ids []string
	appendID := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		ids = append(ids, id)
	}
	for _, id := range req.ActiveAssignmentIDs {
		appendID(id)
	}
	for _, taskID := range req.RunningTaskIDs {
		taskID = strings.TrimSpace(taskID)
		if taskID == "" {
			continue
		}
		for _, assignmentID := range state.agentAssignments[agentID] {
			assignment := state.assignments[assignmentID]
			if assignment.TaskID == taskID && assignmentHoldsActiveLease(assignment.State) {
				appendID(assignment.ID)
			}
		}
	}
	return ids
}
