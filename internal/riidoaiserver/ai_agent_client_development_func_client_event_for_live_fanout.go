package riidoaiserver

func clientEventForLiveFanout(event ClientStreamEvent) (ClientStreamEvent, bool) {
	progressEvent, ok := event.Payload.(AgentThreadProgressEvent)
	if !ok {
		return event, false
	}
	progressEvent.Lines = copyClientVisibleProgressLines(progressEvent.Lines)
	event.Payload = progressEvent
	return event, true
}
