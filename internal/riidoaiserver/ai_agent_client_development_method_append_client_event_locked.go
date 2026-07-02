package riidoaiserver

func (s *DevelopmentAIAgentClientStore) appendClientEventLocked(eventType string, payload any) ClientStreamEvent {
	event := ClientStreamEvent{
		Seq:       s.nextClientEventSeqLocked(),
		EventType: eventType,
		Payload:   payload,
	}
	s.events = appendRetainedClientReplayEvent(s.events, event)
	fanoutEvent, progressFanout := clientEventForLiveFanout(event)
	cache := subscriberEventCache{}
	for _, subscriber := range s.subscribers {
		visible, ok := cache.eventFor(s, subscriber, event, fanoutEvent, progressFanout)
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
