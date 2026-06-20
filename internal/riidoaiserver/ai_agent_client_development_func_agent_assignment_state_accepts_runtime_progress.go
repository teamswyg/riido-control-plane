package riidoaiserver

func agentAssignmentStateAcceptsRuntimeProgress(state AgentAssignmentState) bool {
	switch state {
	case AgentAssignmentStateQueued, AgentAssignmentStateRunning:
		return true
	default:
		return false
	}
}
