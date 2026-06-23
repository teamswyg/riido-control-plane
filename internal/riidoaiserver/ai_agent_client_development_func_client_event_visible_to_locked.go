package riidoaiserver

func clientEventVisibleToLocked(s *DevelopmentAIAgentClientStore, principal AuthorizationResult, event ClientStreamEvent) bool {
	if daemon, ok := eventDeviceDaemon(event.Payload); ok {
		return s.daemonVisibleToPrincipalLocked(principal, daemon)
	}
	if device, ok := eventDeviceRecord(event.Payload); ok {
		return s.deviceVisibleToPrincipalLocked(principal, device)
	}
	agentID, ok := eventAgentID(event.Payload)
	if !ok {
		return true
	}
	agent, exists := s.agents[agentID]
	if !exists {
		taskID, threadID, ok := eventTaskThreadRef(event.Payload)
		if !ok {
			return aiAgentIsAdmin(principal)
		}
		thread, ok := s.visibleTaskThreadLocked(principal, taskID, threadID)
		return ok && thread.AgentID == agentID
	}
	return s.aiAgentVisibleTo(principal, agent)
}
