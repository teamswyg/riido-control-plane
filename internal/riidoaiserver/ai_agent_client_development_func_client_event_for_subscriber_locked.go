package riidoaiserver

func clientEventForSubscriberLocked(
	s *DevelopmentAIAgentClientStore,
	principal AuthorizationResult,
	event ClientStreamEvent,
	fanoutEvent ClientStreamEvent,
	progressFanout bool,
) (ClientStreamEvent, bool) {
	if !progressFanout {
		return clientEventForPrincipalLocked(s, principal, event)
	}
	if !clientEventVisibleToLocked(s, principal, fanoutEvent) {
		return ClientStreamEvent{}, false
	}
	return fanoutEvent, true
}
