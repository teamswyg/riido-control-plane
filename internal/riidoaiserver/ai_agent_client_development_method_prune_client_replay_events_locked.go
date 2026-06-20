package riidoaiserver

func (s *DevelopmentAIAgentClientStore) pruneClientReplayEventsLocked() {
	s.events = retainLatestClientReplayEvents(s.events)
}
