package riidoaiserver

func (s *DevelopmentAIAgentClientStore) clientEventsForPrincipalLocked(principal AuthorizationResult) []ClientStreamEvent {
	events := make([]ClientStreamEvent, 0, len(s.events))
	for _, event := range s.events {
		visible, ok := clientEventForPrincipalLocked(s, principal, event)
		if !ok {
			continue
		}
		events = append(events, visible)
	}
	return withoutSupersededQueuedClientEvents(s, events)
}
