package riidoaiserver

func (s *DevelopmentAIAgentClientStore) appendClientEventLocked(eventType string, payload any) ClientStreamEvent {
	event := ClientStreamEvent{
		Seq:       s.nextClientEventSeqLocked(),
		EventType: eventType,
		Payload:   payload,
	}
	s.events = appendRetainedClientReplayEvent(s.events, event)
	fanoutEvent, progressFanout := clientEventForLiveFanout(event)
	for _, subscriber := range s.subscribers {
		visible, ok := clientEventForSubscriberLocked(s, subscriber.principal, event, fanoutEvent, progressFanout)
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
