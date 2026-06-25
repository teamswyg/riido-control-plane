package riidoaiserver

func copyTaskThreadHistoryMessages(in []AIAgentTaskThreadHistoryMessage) []AIAgentTaskThreadHistoryMessage {
	if len(in) == 0 {
		return nil
	}
	out := make([]AIAgentTaskThreadHistoryMessage, len(in))
	for i, message := range in {
		out[i] = clientVisibleTaskThreadHistoryMessage(message)
	}
	return out
}

func retainLatestThreadHistoryMessages(in []AIAgentTaskThreadHistoryMessage) []AIAgentTaskThreadHistoryMessage {
	if len(in) <= aiAgentClientThreadHistoryMessageLimit {
		return copyTaskThreadHistoryMessages(in)
	}
	return copyTaskThreadHistoryMessages(in[len(in)-aiAgentClientThreadHistoryMessageLimit:])
}
