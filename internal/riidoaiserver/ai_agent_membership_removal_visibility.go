package riidoaiserver

func (s *DevelopmentAIAgentClientStore) taskAgentForMembershipRemovalLocked(principal AuthorizationResult, taskID, agentID, assignmentID string) (AgentClientRecord, bool) {
	agent, ok := s.visibleAgent(principal, agentID)
	if ok {
		return agent, true
	}
	if thread, found := s.taskThreadForUnassignTargetLocked(taskID, agentID, assignmentID); found {
		agent, ok = s.agentFromTaskThreadLocked(principal, thread)
	}
	if !ok && assignmentID == "" {
		if member, found := s.taskAgentMemberships[taskID][agentID]; found && member.RemovedAt == nil {
			agent, ok = s.agentFromTaskThreadLocked(principal, AIAgentTaskThreadRecord{AgentID: member.AgentID, AgentSnapshot: member.AgentSnapshot})
		}
	}
	return agent, ok
}
