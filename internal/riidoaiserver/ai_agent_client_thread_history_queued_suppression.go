package riidoaiserver

func taskThreadHistoryNeedsQueuedSuppression(threads []AIAgentTaskThreadHistoryRecord) bool {
	for _, thread := range threads {
		if taskThreadHistoryThreadIsQueued(thread) {
			return true
		}
		if taskThreadHistoryMessagesHaveQueuedStatus(thread.Messages) {
			return true
		}
	}
	return false
}

func taskThreadHistoryThreadIsQueued(thread AIAgentTaskThreadHistoryRecord) bool {
	return thread.AssignmentState == AgentAssignmentStateQueued ||
		thread.WorkStatus == AgentWorkStatusQueued
}

func taskThreadHistoryMessagesHaveQueuedStatus(
	messages []AIAgentTaskThreadHistoryMessage,
) bool {
	for _, message := range messages {
		if historyMessageIsQueuedStatus(message) {
			return true
		}
	}
	return false
}
