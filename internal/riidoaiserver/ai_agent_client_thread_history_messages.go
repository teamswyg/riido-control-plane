package riidoaiserver

import "time"

func taskThreadProgressMessages(thread AIAgentTaskThreadRecord) []AIAgentTaskThreadHistoryMessage {
	if len(thread.Lines) == 0 {
		return nil
	}
	out := make([]AIAgentTaskThreadHistoryMessage, 0, len(thread.Lines))
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

func taskThreadProgressObservedAt(thread AIAgentTaskThreadRecord, line AgentThreadProgressLine) time.Time {
	if !line.ObservedAt.IsZero() {
		return line.ObservedAt
	}
	return thread.StartedAt
}

func taskThreadProjectionMessage(thread AIAgentTaskThreadRecord) (AIAgentTaskThreadHistoryMessage, bool) {
	if thread.CommentKind == AgentTaskCommentQueuedByBusyAgent {
		return AIAgentTaskThreadHistoryMessage{}, false
	}
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
