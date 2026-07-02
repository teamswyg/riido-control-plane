package riidoaiserver

import "time"

func withoutSupersededQueuedMessages(
	conversationID string,
	messages []AIAgentTaskThreadHistoryMessage,
	latest map[string]time.Time,
	running map[string]struct{},
) []AIAgentTaskThreadHistoryMessage {
	if !taskThreadHistoryMessagesHaveQueuedStatus(messages) {
		return messages
	}
	cutoff, ok := latest[conversationID]
	hasRunning := taskThreadHistoryConversationIsRunning(conversationID, running)
	out := messages[:0]
	for _, message := range messages {
		if historyMessageIsQueuedStatus(message) {
			continue
		}
		if historyQueuedStatusIsSuperseded(message, cutoff, ok, hasRunning) {
			continue
		}
		out = append(out, message)
	}
	return out
}
