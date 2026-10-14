package riidoaiserver

func queuedThreadIsSupersededByRunningConversation(
	thread AIAgentTaskThreadHistoryRecord,
	running map[string]struct{},
) bool {
	if thread.AssignmentState != AgentAssignmentStateQueued &&
		thread.WorkStatus != AgentWorkStatusQueued {
		return false
	}
	return taskThreadHistoryConversationIsRunning(thread.ConversationID, running)
}
