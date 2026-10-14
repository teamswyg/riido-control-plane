package riidoaiserver

func assignmentProgressEvent(input assignmentEventInput, thread AIAgentTaskThreadRecord, line AgentThreadProgressLine) AgentThreadProgressEvent {
	return AgentThreadProgressEvent{
		EventType:       AgentClientEventThreadProgress,
		SchemaVersion:   SchemaVersion,
		AgentID:         input.AgentID,
		TaskID:          thread.TaskID,
		AssignmentID:    input.AssignmentID,
		ThreadID:        thread.ThreadID,
		RunID:           thread.RunID,
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		CommentKind:     AgentTaskCommentRuntimeProgress,
		BatchStartedAt:  line.ObservedAt,
		BatchEndedAt:    line.ObservedAt,
		Lines:           []AgentThreadProgressLine{line},
	}
}
