package riidoaiserver

func newTaskThreadProgressMessageCache(
	thread AIAgentTaskThreadRecord,
	messages []AIAgentTaskThreadHistoryMessage,
) taskThreadProgressMessageCache {
	last := thread.Lines[len(thread.Lines)-1]
	return taskThreadProgressMessageCache{
		assignmentID:       thread.AssignmentID,
		runID:              thread.RunID,
		lineCount:          len(thread.Lines),
		lastSeq:            last.Seq,
		lastMessage:        last.Message,
		lastDisplayMessage: taskProgressDisplayMessage(last),
		lastObserved:       last.ObservedAt,
		visibleLines:       messages,
	}
}
