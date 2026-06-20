package riidoaiserver

func projectedAgentWorkStatusFromActiveThread(thread AIAgentTaskThreadRecord) AgentWorkStatus {
	switch thread.WorkStatus {
	case AgentWorkStatusQueued, AgentWorkStatusRunning, AgentWorkStatusWaitingForUser:
		return thread.WorkStatus
	default:
	}
	switch thread.AssignmentState {
	case AgentAssignmentStateQueued:
		return AgentWorkStatusQueued
	case AgentAssignmentStateRunning, AgentAssignmentStateStopping:
		return AgentWorkStatusRunning
	default:
		return AgentWorkStatusRunning
	}
}
