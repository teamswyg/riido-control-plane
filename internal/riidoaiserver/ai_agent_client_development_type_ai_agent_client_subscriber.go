package riidoaiserver

type aiAgentClientSubscriber struct {
	principal             AuthorizationResult
	visibilityKey         string
	events                chan ClientStreamEvent
	droppedEvents         int64
	terminalCompensations int64
}
