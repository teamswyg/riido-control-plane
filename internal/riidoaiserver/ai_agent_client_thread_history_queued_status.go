package riidoaiserver

func suppressSupersededQueuedHistoryMessages(threads []AIAgentTaskThreadHistoryRecord) {
	if !taskThreadHistoryNeedsQueuedSuppression(threads) {
		return
	}
	running := runningTaskThreadHistoryConversations(threads)
	for i := range threads {
		threads[i].Messages = withoutSupersededQueuedMessages(
			threads[i].Messages,
		)
		if queuedThreadIsSupersededByRunningConversation(threads[i], running) {
			threads[i].ActiveStream = nil
		}
	}
}

func historyMessageIsQueuedStatus(message AIAgentTaskThreadHistoryMessage) bool {
	return message.Role == AIAgentTaskThreadMessageRoleAgent &&
		message.CommentKind == AgentTaskCommentQueuedByBusyAgent
}
