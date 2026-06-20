package riidoaiserver

func (s *DevelopmentAIAgentClientStore) appendClientEventLocked(eventType string, payload any) ClientStreamEvent {
	event := ClientStreamEvent{
		Seq:       s.nextClientEventSeqLocked(),
		EventType: eventType,
		Payload:   payload,
	}
	s.events = append(s.events, event)
	s.pruneClientReplayEventsLocked()
	for _, subscriber := range s.subscribers {
		visible, ok := clientEventForPrincipalLocked(s, subscriber.principal, event)
		if !ok {
			continue
		}
		select {
		case subscriber.events <- visible:
		default:
		}
	}
	return event
}
