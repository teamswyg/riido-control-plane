package riidoaiserver

func fallbackAgentThreadProgressResponse(agentID string, req AgentThreadProgressBatchRequest) AgentThreadProgressBatchResponse {
	threadID := req.ThreadID
	if threadID == "" {
		threadID = threadIDForRun(req.TaskID, agentID, req.RunID)
	}
	return AgentThreadProgressBatchResponse{
		SchemaVersion: SchemaVersion,
		AcceptedLines: len(req.Lines),
		Event: AgentThreadProgressEvent{
			EventType:       AgentClientEventThreadProgress,
			SchemaVersion:   SchemaVersion,
			AgentID:         agentID,
			TaskID:          req.TaskID,
			ThreadID:        threadID,
			RunID:           req.RunID,
			WorkStatus:      AgentWorkStatusRunning,
			AssignmentState: AgentAssignmentStateRunning,
			CommentKind:     AgentTaskCommentRuntimeProgress,
			BatchStartedAt:  req.BatchStartedAt,
			BatchEndedAt:    req.BatchEndedAt,
			Lines:           req.Lines,
		},
	}
}
