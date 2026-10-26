package riidoaiserver

// agentContextReadEventForPrincipalLocked removes agent-directory sources the
// subscriber cannot see now (RIID-5900 1-b). A run may list the requester's
// private agents or, for an admin requester, every workspace agent; other
// viewers of the same thread must not learn those agent IDs or profile hashes.
// Hidden sources are dropped and excluded from omitted_sources, so their
// number is not revealed either.
func (s *DevelopmentAIAgentClientStore) agentContextReadEventForPrincipalLocked(principal AuthorizationResult, event AgentContextReadEvent) AgentContextReadEvent {
	visible := func(id string) bool {
		agent, exists := s.agents[id]
		return exists && s.aiAgentVisibleTo(principal, agent)
	}
	sources := make([]AgentContextReadSource, 0, len(event.Sources))
	directory := len(event.omittedAgentIDs) > 0
	for _, source := range event.Sources {
		if source.Kind == "agent" {
			directory = true
			if !visible(source.ID) {
				continue
			}
		}
		sources = append(sources, source)
	}
	event.Sources = sources
	switch {
	case event.omittedOther+len(event.omittedAgentIDs) == event.OmittedSources:
		// Fresh event: count only cut-off sources this subscriber could see.
		event.OmittedSources = event.omittedOther
		for _, id := range event.omittedAgentIDs {
			if visible(id) {
				event.OmittedSources++
			}
		}
	case directory:
		// Restored without bookkeeping: a count could include hidden agents.
		event.OmittedSources = 0
	}
	event.omittedAgentIDs, event.omittedOther = nil, 0
	return event
}

// noteOmitted records a source cut by the per-event cap.
func (e *AgentContextReadEvent) noteOmitted(read AssignmentContextRead) {
	e.OmittedSources++
	if read.SourceKind == "agent" {
		e.omittedAgentIDs = append(e.omittedAgentIDs, read.SourceID)
		return
	}
	e.omittedOther++
}
