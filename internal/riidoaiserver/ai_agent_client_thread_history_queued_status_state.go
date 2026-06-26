package riidoaiserver

import "time"

func threadStatusSupersedesQueuedStatus(thread AIAgentTaskThreadHistoryRecord) (time.Time, bool) {
	if !threadIsActivelyRunning(thread) {
		return time.Time{}, false
	}
	observedAt := thread.StartedAt
	if observedAt.IsZero() {
		observedAt = thread.CompletedAt
	}
	return observedAt, !observedAt.IsZero()
}

func threadIsActivelyRunning(thread AIAgentTaskThreadHistoryRecord) bool {
	return thread.WorkStatus == AgentWorkStatusRunning ||
		thread.AssignmentState == AgentAssignmentStateRunning
}
