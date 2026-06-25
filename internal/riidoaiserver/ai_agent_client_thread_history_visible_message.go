package riidoaiserver

func clientVisibleTaskThreadHistoryMessage(message AIAgentTaskThreadHistoryMessage) AIAgentTaskThreadHistoryMessage {
	switch message.Role {
	case AIAgentTaskThreadMessageRoleAgent, AIAgentTaskThreadMessageRoleProgress:
		message.Body = clientVisibleTaskThreadText(message.Body)
		message.ResultMessage = clientVisibleTaskThreadText(message.ResultMessage)
	case AIAgentTaskThreadMessageRoleUser:
	}
	return message
}
