package riidoaiserver

func taskThreadMessageStartsNewExecution(thread AIAgentTaskThreadRecord) bool {
	switch thread.AssignmentState {
	case AgentAssignmentStateCompleted, AgentAssignmentStateFailed,
		AgentAssignmentStateStopped, AgentAssignmentStateUnassigned:
		return true
	default:
		return false
	}
}
