package riidoaiserver

func terminalProgressEventFromThread(thread AIAgentTaskThreadRecord) (AgentThreadProgressEvent, bool) {
	if !agentAssignmentStateIsTerminal(thread.AssignmentState) {
		return AgentThreadProgressEvent{}, false
	}
	return AgentThreadProgressEvent{
		EventType:       AgentClientEventThreadProgress,
		SchemaVersion:   SchemaVersion,
		AgentID:         thread.AgentID,
		TaskID:          thread.TaskID,
		AssignmentID:    thread.AssignmentID,
		ThreadID:        thread.ThreadID,
		RunID:           thread.RunID,
		WorkStatus:      thread.WorkStatus,
		AssignmentState: thread.AssignmentState,
		CommentKind:     thread.CommentKind,
		Lines:           []AgentThreadProgressLine{},
	}, true
}
