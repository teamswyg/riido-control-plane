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
	var out []AIAgentTaskThreadHistoryMessage
	for i, message := range messages {
		if historyThreadMessageShouldDropQueued(message, cutoff, ok, hasRunning) {
			if out == nil {
				out = make([]AIAgentTaskThreadHistoryMessage, 0, len(messages)-1)
				out = append(out, messages[:i]...)
			}
			continue
		}
		if out != nil {
			out = append(out, message)
		}
	}
	if out == nil {
		return messages
	}
	return out
}

func historyThreadMessageShouldDropQueued(
	message AIAgentTaskThreadHistoryMessage,
	cutoff time.Time,
	hasCutoff bool,
	hasRunning bool,
) bool {
	return historyMessageIsQueuedStatus(message) ||
		historyQueuedStatusIsSuperseded(message, cutoff, hasCutoff, hasRunning)
}
