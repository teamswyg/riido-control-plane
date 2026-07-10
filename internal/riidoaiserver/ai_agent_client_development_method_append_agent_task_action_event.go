package riidoaiserver

func (s *DevelopmentAIAgentClientStore) appendAgentTaskActionEvent(response AIAgentTaskActionResponse) {
	s.appendClientEventLocked(AgentClientEventWorkStatusChanged, AgentWorkStatusChangedEvent{
		EventType:          AgentClientEventWorkStatusChanged,
		SchemaVersion:      SchemaVersion,
		AgentID:            response.AgentID,
		TaskID:             response.TaskID,
		AssignmentID:       response.AssignmentID,
		ThreadID:           response.ThreadID,
		RunID:              response.RunID,
		WorkStatus:         response.WorkStatus,
		AssignmentState:    response.AssignmentState,
		CommentKind:        response.CommentKind,
		ResultMessage:      response.ResultMessage,
		FailureDiagnostics: copyFailureDiagnostics(response.FailureDiagnostics),
	})
	if event, ok := terminalAgentTaskProgressEvent(response); ok {
		s.appendClientEventLocked(event.EventType, event)
	}
}
