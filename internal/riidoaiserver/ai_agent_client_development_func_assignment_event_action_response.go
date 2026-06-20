package riidoaiserver

func assignmentEventActionResponse(thread AIAgentTaskThreadRecord, state AssignmentState, message string, metadata map[string]string) AIAgentTaskActionResponse {
	response := AIAgentTaskActionResponse{
		SchemaVersion:   SchemaVersion,
		TaskID:          thread.TaskID,
		AssignmentID:    thread.AssignmentID,
		AgentID:         thread.AgentID,
		ThreadID:        thread.ThreadID,
		RunID:           thread.RunID,
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		CommentKind:     AgentTaskCommentRuntimeProgress,
		Message:         clientVisibleTaskThreadText(message),
	}
	applyAssignmentStateActionResponse(&response, state)
	if assignmentStateCarriesResultMessage(state) {
		response.ResultMessage = response.Message
	}
	if response.AssignmentState == AgentAssignmentStateFailed {
		response.FailureDiagnostics = failureDiagnosticsFromAssignmentEvent(metadata, response.Message)
	}
	return response
}
