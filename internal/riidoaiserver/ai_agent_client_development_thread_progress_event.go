package riidoaiserver

func (input threadProgressInput) event() AgentThreadProgressEvent {
	return AgentThreadProgressEvent{
		EventType:       AgentClientEventThreadProgress,
		SchemaVersion:   SchemaVersion,
		AgentID:         input.AgentID,
		TaskID:          input.Request.TaskID,
		AssignmentID:    input.Request.AssignmentID,
		ThreadID:        input.Request.ThreadID,
		RunID:           input.Request.RunID,
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		CommentKind:     AgentTaskCommentRuntimeProgress,
		BatchStartedAt:  input.Request.BatchStartedAt,
		BatchEndedAt:    input.Request.BatchEndedAt,
		Lines:           input.Lines,
	}
}

func (input threadProgressInput) noopResponse() AgentThreadProgressBatchResponse {
	event := input.event()
	event.Lines = nil
	return AgentThreadProgressBatchResponse{
		SchemaVersion: SchemaVersion,
		AcceptedLines: 0,
		Event:         event,
	}
}
