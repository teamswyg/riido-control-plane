package riidoaiserver

func taskThreadMessageStartsNewExecution(thread AIAgentTaskThreadRecord) bool {
	if thread.WorkStatus == AgentWorkStatusWaitingForUser {
		return true
	}
	switch thread.AssignmentState {
	case AgentAssignmentStateCompleted, AgentAssignmentStateFailed,
		AgentAssignmentStateStopped, AgentAssignmentStateUnassigned:
		return true
	default:
		return false
	}
}
