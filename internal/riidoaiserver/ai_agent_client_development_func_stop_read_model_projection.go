package riidoaiserver

func stopReadModelProjection(durable AssignmentState) (AgentWorkStatus, AgentAssignmentState, bool) {
	switch durable.Code() {
	case AssignmentStateCodeCancelling:
		return AgentWorkStatusIdle, AgentAssignmentStateStopped, true
	case AssignmentStateCodeCancelled:
		return AgentWorkStatusIdle, AgentAssignmentStateStopped, true
	default:
		return AgentWorkStatusIdle, AgentAssignmentStateStopped, true
	}
}
