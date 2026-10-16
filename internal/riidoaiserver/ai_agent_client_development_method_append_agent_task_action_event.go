package riidoaiserver

func (s *DevelopmentAIAgentClientStore) appendAgentTaskActionEvent(response AIAgentTaskActionResponse) {
	s.appendAgentTaskActionEventWithFanout(response, true)
}

func (s *DevelopmentAIAgentClientStore) appendAgentTaskActionEventWithFanout(response AIAgentTaskActionResponse, fanout bool) {
	appendEvent := s.appendClientEventWithoutFanoutLocked
	if fanout {
		appendEvent = s.appendClientEventLocked
	}
	appendEvent(AgentClientEventWorkStatusChanged, AgentWorkStatusChangedEvent{
		EventType:          AgentClientEventWorkStatusChanged,
		SchemaVersion:      SchemaVersion,
		AgentID:            response.AgentID,
		TaskID:             response.TaskID,
		AssignmentID:       response.AssignmentID,
		ThreadID:           response.ThreadID,
		RunID:              response.RunID,
		WorkStatus:         response.WorkStatus,
		ConfirmationReason: response.ConfirmationReason,
		AssignmentState:    response.AssignmentState,
		CommentKind:        response.CommentKind,
		ResultMessage:      response.ResultMessage,
		FailureDiagnostics: copyFailureDiagnostics(response.FailureDiagnostics),
		ExecutionContext:   copyExecutionContext(response.ExecutionContext),
	})
	if event, ok := terminalAgentTaskProgressEvent(response); ok {
		appendEvent(event.EventType, event)
	}
}
