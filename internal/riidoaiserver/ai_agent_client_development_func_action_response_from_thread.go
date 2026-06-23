package riidoaiserver

func actionResponseFromThread(thread AIAgentTaskThreadRecord, workspaceID string) AIAgentTaskActionResponse {
	response := AIAgentTaskActionResponse{
		SchemaVersion:      SchemaVersion,
		TaskID:             thread.TaskID,
		AssignmentID:       thread.AssignmentID,
		AgentID:            thread.AgentID,
		AgentSnapshot:      copyTaskThreadAgentSnapshot(thread.AgentSnapshot),
		ThreadID:           thread.ThreadID,
		RunID:              thread.RunID,
		WorkStatus:         thread.WorkStatus,
		AssignmentState:    thread.AssignmentState,
		CommentKind:        thread.CommentKind,
		Message:            clientVisibleTaskThreadMessage(thread),
		ResultMessage:      clientVisibleTaskThreadResultMessage(thread),
		FailureDiagnostics: clientVisibleFailureDiagnostics(thread.FailureDiagnostics),
	}
	return actionResponseWithActiveStream(response, workspaceID)
}
