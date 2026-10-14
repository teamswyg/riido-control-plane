package riidoaiserver

import (
	"strings"
)

func assignmentProjectionMatchesTaskThread(thread AIAgentTaskThreadRecord, projection AssignmentProjection) bool {
	assignment := projection.Assignment
	if strings.TrimSpace(assignment.ID) == "" ||
		strings.TrimSpace(assignment.ID) != strings.TrimSpace(thread.AssignmentID) ||
		strings.TrimSpace(assignment.AgentID) != strings.TrimSpace(thread.AgentID) {
		return false
	}
	taskID := strings.TrimSpace(thread.TaskID)
	if taskID == "" {
		return true
	}
	return strings.TrimSpace(assignment.TaskID) == taskID ||
		strings.TrimSpace(assignment.ComponentID) == taskID
}
