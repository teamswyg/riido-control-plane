package riidoaiserver

func taskThreadProgressMessages(thread AIAgentTaskThreadRecord) []AIAgentTaskThreadHistoryMessage {
	if len(thread.Lines) == 0 {
		return nil
	}
	out := make([]AIAgentTaskThreadHistoryMessage, 0, len(thread.Lines))
	for _, line := range thread.Lines {
		if line.Message == "" {
			continue
		}
		out = append(out, AIAgentTaskThreadHistoryMessage{
			MessageID:    taskThreadProgressMessageID(thread.ThreadID, line.Seq),
			Role:         AIAgentTaskThreadMessageRoleProgress,
			CommentKind:  AgentTaskCommentRuntimeProgress,
			AssignmentID: thread.AssignmentID,
			RunID:        thread.RunID,
			Seq:          line.Seq,
			Body:         line.Message,
			ObservedAt:   line.ObservedAt,
		})
	}
	return out
}

func taskThreadProjectionMessage(thread AIAgentTaskThreadRecord) (AIAgentTaskThreadHistoryMessage, bool) {
	body := clientVisibleTaskThreadMessage(thread)
	result := clientVisibleTaskThreadResultMessage(thread)
	if body == "" && result == "" {
		return AIAgentTaskThreadHistoryMessage{}, false
	}
	observedAt := thread.CompletedAt
	if observedAt.IsZero() {
		observedAt = thread.StartedAt
	}
	return AIAgentTaskThreadHistoryMessage{
		MessageID:     taskThreadProjectionMessageID(thread),
		Role:          AIAgentTaskThreadMessageRoleAgent,
		CommentKind:   thread.CommentKind,
		AssignmentID:  thread.AssignmentID,
		RunID:         thread.RunID,
		Body:          body,
		ResultMessage: result,
		ObservedAt:    observedAt,
	}, true
}
