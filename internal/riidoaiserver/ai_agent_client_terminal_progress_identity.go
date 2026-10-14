package riidoaiserver

type terminalProgressIdentity struct {
	taskID, agentID, assignmentID, threadID, runID string
	state                                          AgentAssignmentState
}

func clientStreamEventIsTerminalProgress(event ClientStreamEvent) bool {
	progress, ok := event.Payload.(AgentThreadProgressEvent)
	return ok && agentAssignmentStateIsTerminal(progress.AssignmentState)
}

func agentAssignmentStateIsTerminal(state AgentAssignmentState) bool {
	switch state {
	case AgentAssignmentStateCompleted,
		AgentAssignmentStateFailed,
		AgentAssignmentStateStopped,
		AgentAssignmentStateUnassigned:
		return true
	default:
		return false
	}
}

func terminalProgressMatchesThread(event ClientStreamEvent, thread AIAgentTaskThreadRecord) bool {
	eventIdentity, ok := terminalProgressIdentityFromEvent(event)
	return ok && eventIdentity == terminalProgressIdentityFromThread(thread)
}

func terminalProgressIdentityFromEvent(event ClientStreamEvent) (terminalProgressIdentity, bool) {
	progress, ok := event.Payload.(AgentThreadProgressEvent)
	if !ok || !agentAssignmentStateIsTerminal(progress.AssignmentState) {
		return terminalProgressIdentity{}, false
	}
	return terminalProgressIdentity{
		taskID: progress.TaskID, agentID: progress.AgentID, assignmentID: progress.AssignmentID,
		threadID: progress.ThreadID, runID: progress.RunID, state: progress.AssignmentState,
	}, true
}

func terminalProgressIdentityFromThread(thread AIAgentTaskThreadRecord) terminalProgressIdentity {
	return terminalProgressIdentity{
		taskID: thread.TaskID, agentID: thread.AgentID, assignmentID: thread.AssignmentID,
		threadID: thread.ThreadID, runID: thread.RunID, state: thread.AssignmentState,
	}
}
