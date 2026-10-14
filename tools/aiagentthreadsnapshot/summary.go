package main

func summarizeThreads(payload threadCollection, conversationID string) threadSummary {
	summary := threadSummary{
		ThreadCount:  len(payload.Threads),
		ActiveStream: len(payload.ActiveStream) > 0,
	}
	for _, thread := range payload.Threads {
		if isRunning(thread) {
			summary.RunningCount++
		}
		if thread.WorkStatus == "queued" || thread.AssignmentState == "queued" {
			summary.QueuedCount++
		}
		if isTerminal(thread.AssignmentState) {
			summary.TerminalCount++
			if len(thread.ActiveStream) > 0 {
				summary.TerminalActiveStreamCount++
			}
		}
		if shouldHighlight(thread, conversationID) {
			summary.HighlightedThreads = append(summary.HighlightedThreads, surface(thread))
		}
	}
	return summary
}

func shouldHighlight(thread threadRecord, conversationID string) bool {
	return conversationID == "" || thread.ConversationID == conversationID ||
		thread.ThreadID == conversationID
}

func surface(thread threadRecord) threadSurface {
	return threadSurface{
		ThreadID: thread.ThreadID, ConversationID: thread.ConversationID,
		AssignmentID: thread.AssignmentID, RunID: thread.RunID,
		WorkStatus: thread.WorkStatus, AssignmentState: thread.AssignmentState,
		CommentKind: thread.CommentKind, MessageCount: len(thread.Messages),
		LineCount: len(thread.Lines), ActiveStream: len(thread.ActiveStream) > 0,
	}
}

func isRunning(thread threadRecord) bool {
	return thread.WorkStatus == "running" || thread.AssignmentState == "running"
}
