package riidoaiserver

func threadIsActivelyRunning(thread AIAgentTaskThreadHistoryRecord) bool {
	return thread.WorkStatus == AgentWorkStatusRunning ||
		thread.AssignmentState == AgentAssignmentStateRunning
}
