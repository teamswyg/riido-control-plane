package riidoaiserver

func queueDiagnosticsEqual(a, b *AIAgentTaskThreadQueueDiagnostics) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Reason == b.Reason &&
		a.BlockedByAssignmentID == b.BlockedByAssignmentID &&
		a.BlockerAgentID == b.BlockerAgentID &&
		a.BlockerRuntimeProvider == b.BlockerRuntimeProvider &&
		a.BlockerState == b.BlockerState &&
		a.BlockerUpdatedAt.Equal(b.BlockerUpdatedAt)
}
