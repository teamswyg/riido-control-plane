package riidoaiserver

import (
	"strings"
)

func assignmentEventVisibleThreadMessage(state AssignmentState, eventType, message, previous string) string {
	switch state.Code() {
	case AssignmentStateCodeQueued, AssignmentStateCodeLeased, AssignmentStateCodeReady, AssignmentStateCodeRunning, AssignmentStateCodeCancelling:
		if strings.TrimSpace(eventType) != EventRiidoLog {
			return strings.TrimSpace(previous)
		}
	default:
	}
	return strings.TrimSpace(message)
}
