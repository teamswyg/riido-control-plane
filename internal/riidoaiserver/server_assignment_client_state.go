package riidoaiserver

import "strings"

func assignmentClientResponseDurableState(assignment Assignment) AssignmentState {
	if strings.TrimSpace(assignment.BlockedByAssignmentID) != "" {
		return assignment.State
	}
	switch assignment.State.Code() {
	case AssignmentStateCodeReady, AssignmentStateCodeRunning, AssignmentStateCodeCancelling,
		AssignmentStateCodeCancelled, AssignmentStateCodeCompleted, AssignmentStateCodeFailed:
		return assignment.State
	default:
		return ""
	}
}
