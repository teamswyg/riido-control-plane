package riidoaiserver

import "time"

func suppressSupersededQueuedHistoryMessages(threads []AIAgentTaskThreadHistoryRecord) {
	latest := latestConversationNonQueuedMessageTime(threads)
	for i := range threads {
		threads[i].Messages = withoutSupersededQueuedMessages(threads[i].ConversationID, threads[i].Messages, latest)
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

func withoutSupersededQueuedMessages(
	conversationID string,
	messages []AIAgentTaskThreadHistoryMessage,
	latest map[string]time.Time,
) []AIAgentTaskThreadHistoryMessage {
	cutoff, ok := latest[conversationID]
	if !ok {
		return messages
	}
	out := messages[:0]
	for _, message := range messages {
		if historyMessageIsQueuedStatus(message) && !cutoff.Before(message.ObservedAt) {
			continue
		}
		out = append(out, message)
	}
	return out
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
