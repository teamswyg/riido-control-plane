package riidoaiserver

func taskThreadCarriesResultMessage(thread AIAgentTaskThreadRecord) bool {
	switch thread.AssignmentState {
	case AgentAssignmentStateCompleted, AgentAssignmentStateFailed, AgentAssignmentStateStopped:
		return true
	default:
		return false
	}
}
