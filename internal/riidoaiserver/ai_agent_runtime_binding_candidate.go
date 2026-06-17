package riidoaiserver

type agentRuntimeBindingCandidate struct {
	binding AgentRuntimeBinding
	agent   AgentClientRecord
}

func preferRuntimeBindingCandidate(next, current agentRuntimeBindingCandidate) bool {
	nextActive := agentRuntimeBindingCandidateActive(next)
	currentActive := agentRuntimeBindingCandidateActive(current)
	if nextActive != currentActive {
		return nextActive
	}
	if !next.agent.UpdatedAt.Equal(current.agent.UpdatedAt) {
		return next.agent.UpdatedAt.After(current.agent.UpdatedAt)
	}
	if !next.agent.CreatedAt.Equal(current.agent.CreatedAt) {
		return next.agent.CreatedAt.After(current.agent.CreatedAt)
	}
	return next.agent.AgentID < current.agent.AgentID
}

func agentRuntimeBindingCandidateActive(candidate agentRuntimeBindingCandidate) bool {
	return candidate.agent.AssignedTaskCount > 0 ||
		candidate.agent.WorkStatus == AgentWorkStatusQueued ||
		candidate.agent.WorkStatus == AgentWorkStatusRunning ||
		candidate.agent.WorkStatus == AgentWorkStatusWaitingForUser
}
