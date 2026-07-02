package riidoaiserver

import "time"

func suppressSupersededQueuedHistoryMessages(threads []AIAgentTaskThreadHistoryRecord) {
	if !taskThreadHistoryNeedsQueuedSuppression(threads) {
		return
	}
	latest := latestConversationNonQueuedMessageTime(threads)
	running := runningTaskThreadHistoryConversations(threads)
	for i := range threads {
		threads[i].Messages = withoutSupersededQueuedMessages(
			threads[i].ConversationID, threads[i].Messages, latest, running,
		)
		if queuedThreadIsSupersededByRunningConversation(threads[i], running) {
			threads[i].ActiveStream = nil
		}
	}
}

func latestConversationNonQueuedMessageTime(threads []AIAgentTaskThreadHistoryRecord) map[string]time.Time {
	latest := map[string]time.Time{}
	for _, thread := range threads {
		for _, message := range thread.Messages {
			if !historyMessageSupersedesQueuedStatus(message) {
				continue
			}
			if latest[thread.ConversationID].Before(message.ObservedAt) {
				latest[thread.ConversationID] = message.ObservedAt
			}
		}
		if observedAt, ok := threadStatusSupersedesQueuedStatus(thread); ok &&
			latest[thread.ConversationID].Before(observedAt) {
			latest[thread.ConversationID] = observedAt
		}
	}
	return latest
}

func historyMessageIsQueuedStatus(message AIAgentTaskThreadHistoryMessage) bool {
	return message.Role == AIAgentTaskThreadMessageRoleAgent &&
		message.CommentKind == AgentTaskCommentQueuedByBusyAgent
}

func historyMessageSupersedesQueuedStatus(message AIAgentTaskThreadHistoryMessage) bool {
	if message.Role == AIAgentTaskThreadMessageRoleProgress {
		return true
	}
	return message.Role == AIAgentTaskThreadMessageRoleAgent &&
		message.CommentKind != "" &&
		message.CommentKind != AgentTaskCommentQueuedByBusyAgent
}
