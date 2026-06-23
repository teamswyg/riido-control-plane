package riidoaiserver

func copyTaskThreadHistoryMessages(in []AIAgentTaskThreadHistoryMessage) []AIAgentTaskThreadHistoryMessage {
	if len(in) == 0 {
		return nil
	}
	return append([]AIAgentTaskThreadHistoryMessage(nil), in...)
}

func retainLatestThreadHistoryMessages(in []AIAgentTaskThreadHistoryMessage) []AIAgentTaskThreadHistoryMessage {
	if len(in) <= aiAgentClientThreadHistoryMessageLimit {
		return copyTaskThreadHistoryMessages(in)
	}
	return copyTaskThreadHistoryMessages(in[len(in)-aiAgentClientThreadHistoryMessageLimit:])
}
