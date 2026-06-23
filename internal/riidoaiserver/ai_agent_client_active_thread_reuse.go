package riidoaiserver

import "strings"

func canReuseActiveTaskThreadForAssignment(thread AIAgentTaskThreadRecord, assignmentID string) bool {
	assignmentID = strings.TrimSpace(assignmentID)
	if assignmentID == "" {
		return true
	}
	return assignmentID == strings.TrimSpace(thread.AssignmentID)
}
