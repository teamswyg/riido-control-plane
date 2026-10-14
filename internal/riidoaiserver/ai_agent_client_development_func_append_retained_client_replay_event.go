package riidoaiserver

func appendRetainedClientReplayEvent(events []ClientStreamEvent, event ClientStreamEvent) []ClientStreamEvent {
	if len(events) < aiAgentClientReplayEventLimit {
		return append(events, event)
	}
	if len(events) == aiAgentClientReplayEventLimit {
		copy(events, events[1:])
		events[len(events)-1] = event
		return events
	}
	retained := retainLatestClientReplayEvents(events)
	return appendRetainedClientReplayEvent(retained, event)
}
