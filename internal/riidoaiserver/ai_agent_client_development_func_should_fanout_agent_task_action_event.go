package riidoaiserver

func shouldFanoutAgentTaskActionEvent(hadThread bool, previous AIAgentTaskThreadRecord, response AIAgentTaskActionResponse) bool {
	if !hadThread {
		return true
	}
	return previous.WorkStatus != response.WorkStatus ||
		previous.AssignmentState != response.AssignmentState ||
		previous.CommentKind != response.CommentKind ||
		!failureDiagnosticsEqual(previous.FailureDiagnostics, response.FailureDiagnostics)
}
