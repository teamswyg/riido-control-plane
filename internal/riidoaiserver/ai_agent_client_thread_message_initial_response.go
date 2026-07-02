package riidoaiserver

func taskThreadMessageInitialResponse(
	taskID string,
	threadID string,
	thread AIAgentTaskThreadRecord,
	agent AgentClientRecord,
) AIAgentTaskActionResponse {
	return AIAgentTaskActionResponse{
		SchemaVersion:   SchemaVersion,
		TaskID:          taskID,
		AssignmentID:    thread.AssignmentID,
		AgentID:         agent.AgentID,
		AgentSnapshot:   copyTaskThreadAgentSnapshot(thread.AgentSnapshot),
		ThreadID:        threadID,
		RunID:           thread.RunID,
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		CommentKind:     AgentTaskCommentRuntimeProgress,
		Message:         clientMessageTaskRunning,
	}
}
