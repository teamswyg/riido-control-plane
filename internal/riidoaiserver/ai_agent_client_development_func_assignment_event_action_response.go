package riidoaiserver

func assignmentEventActionResponse(thread AIAgentTaskThreadRecord, state AssignmentState, message string, metadata map[string]string) AIAgentTaskActionResponse {
	response := AIAgentTaskActionResponse{
		SchemaVersion:   SchemaVersion,
		TaskID:          thread.TaskID,
		AssignmentID:    thread.AssignmentID,
		AgentID:         thread.AgentID,
		AgentSnapshot:   copyTaskThreadAgentSnapshot(thread.AgentSnapshot),
		ThreadID:        thread.ThreadID,
		RunID:           thread.RunID,
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		CommentKind:     AgentTaskCommentRuntimeProgress,
		Message:         clientVisibleFailureMessage(metadata, message),
	}
	applyAssignmentStateActionResponse(&response, state)
	applyAssignmentMetadataActionResponse(&response, metadata)
	if assignmentStateCarriesResultMessage(state) {
		response.ResultMessage = response.Message
	}
	if response.AssignmentState == AgentAssignmentStateFailed {
		response.FailureDiagnostics = failureDiagnosticsFromAssignmentEvent(metadata, message)
	}
	return response
}
