package riidoaiserver

func withoutSupersededQueuedMessages(
	messages []AIAgentTaskThreadHistoryMessage,
) []AIAgentTaskThreadHistoryMessage {
	if !taskThreadHistoryMessagesHaveQueuedStatus(messages) {
		return messages
	}
	var out []AIAgentTaskThreadHistoryMessage
	for i, message := range messages {
		if historyThreadMessageShouldDropQueued(message) {
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
