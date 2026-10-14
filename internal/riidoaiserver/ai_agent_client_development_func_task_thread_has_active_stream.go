package riidoaiserver

func taskThreadHasActiveStream(thread AIAgentTaskThreadRecord) bool {
	switch thread.AssignmentState {
	case AgentAssignmentStateQueued, AgentAssignmentStateRunning:
		return true
	default:
		return false
	}
}

// agentAssignmentStateAcceptsRuntimeProgress reports whether a thread in this
// state may be advanced by an incoming runtime-progress event (a `riido_log`
// on /events, or a /thread-progress batch). Only genuinely active states accept
// progress; once a thread is stopped/terminal, late progress — e.g. an
// in-flight log that raced a user Stop — must NOT re-activate it (flip it back
// to running, re-open the SSE active_stream, or re-arm the agent). This is the
// control-plane read-model terminal/stop fence (C2).
