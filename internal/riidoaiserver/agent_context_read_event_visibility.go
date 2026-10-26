package riidoaiserver

// agentContextReadEventForPrincipalLocked removes agent-directory sources the
// subscriber cannot see now (RIID-5900 1-b). A run may list the requester's
// private agents or, for an admin requester, every workspace agent; other
// viewers of the same thread must not learn those agent IDs or profile hashes.
// Hidden sources are dropped without being counted, so their number is not
// revealed either.
func (s *DevelopmentAIAgentClientStore) agentContextReadEventForPrincipalLocked(principal AuthorizationResult, event AgentContextReadEvent) AgentContextReadEvent {
	sources := make([]AgentContextReadSource, 0, len(event.Sources))
	for _, source := range event.Sources {
		if source.Kind == "agent" {
			agent, exists := s.agents[source.ID]
			if !exists || !s.aiAgentVisibleTo(principal, agent) {
				continue
			}
		}
		sources = append(sources, source)
	}
	event.Sources = sources
	return event
}
