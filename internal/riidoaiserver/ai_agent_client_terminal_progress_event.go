package riidoaiserver

func terminalAgentTaskProgressEvent(response AIAgentTaskActionResponse) (AgentThreadProgressEvent, bool) {
	if !agentTaskActionIsTerminal(response) {
		return AgentThreadProgressEvent{}, false
	}
	return AgentThreadProgressEvent{
		EventType:       AgentClientEventThreadProgress,
		SchemaVersion:   SchemaVersion,
		AgentID:         response.AgentID,
		TaskID:          response.TaskID,
		AssignmentID:    response.AssignmentID,
		ThreadID:        response.ThreadID,
		RunID:           response.RunID,
		WorkStatus:      response.WorkStatus,
		AssignmentState: response.AssignmentState,
		CommentKind:     response.CommentKind,
		Lines:           []AgentThreadProgressLine{},
	}, true
}

func agentTaskActionIsTerminal(response AIAgentTaskActionResponse) bool {
	return agentAssignmentStateIsTerminal(response.AssignmentState)
}
