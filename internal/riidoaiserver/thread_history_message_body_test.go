package riidoaiserver

func historyMessagesContainBody(messages []AIAgentTaskThreadHistoryMessage, body string) bool {
	for _, message := range messages {
		if message.Body == body || message.ResultMessage == body {
			return true
		}
	}
	return false
}
