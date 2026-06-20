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
		return aiAgentIsAdmin(principal)
	}
	return s.aiAgentVisibleTo(principal, agent)
}
