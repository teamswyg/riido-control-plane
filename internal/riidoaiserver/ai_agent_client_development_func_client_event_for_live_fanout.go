package riidoaiserver

func clientEventForLiveFanout(event ClientStreamEvent) (ClientStreamEvent, bool) {
	progressEvent, ok := event.Payload.(AgentThreadProgressEvent)
	if ok {
		progressEvent.Lines = copyClientVisibleProgressLines(progressEvent.Lines)
		event.Payload = progressEvent
		return event, true
	}
	if statusEvent, ok := event.Payload.(AgentWorkStatusChangedEvent); ok {
		event.Payload = clientVisibleWorkStatusChangedEvent(statusEvent)
	}
	return event, false
}
