package riidoaiserver

func buildTaskThreadHistoryMessages(
	stored []AIAgentTaskThreadHistoryMessage,
	thread AIAgentTaskThreadRecord,
) []AIAgentTaskThreadHistoryMessage {
	size := len(stored) + len(thread.Lines) + 1
	if size == 1 {
		size = 0
	}
	out := make([]AIAgentTaskThreadHistoryMessage, 0, size)
	out = appendVisibleStoredTaskThreadHistoryMessages(out, stored)
	return appendTaskThreadProgressMessages(out, thread)
}

func appendVisibleStoredTaskThreadHistoryMessages(
	out []AIAgentTaskThreadHistoryMessage,
	stored []AIAgentTaskThreadHistoryMessage,
) []AIAgentTaskThreadHistoryMessage {
	for _, message := range stored {
		out = append(out, clientVisibleTaskThreadHistoryMessage(message))
	}
	return out
}

func appendTaskThreadProgressMessages(
	out []AIAgentTaskThreadHistoryMessage,
	thread AIAgentTaskThreadRecord,
) []AIAgentTaskThreadHistoryMessage {
	for _, line := range thread.Lines {
		body := clientVisibleTaskThreadText(line.Message)
		if body == "" {
			continue
		}
		out = append(out, AIAgentTaskThreadHistoryMessage{
			MessageID:    taskThreadProgressMessageID(thread.ThreadID, line.Seq),
			Role:         AIAgentTaskThreadMessageRoleProgress,
			CommentKind:  AgentTaskCommentRuntimeProgress,
			AssignmentID: thread.AssignmentID,
			RunID:        thread.RunID,
			Seq:          line.Seq,
			Body:         body,
			ObservedAt:   taskThreadProgressObservedAt(thread, line),
		})
	}
	return out
}
