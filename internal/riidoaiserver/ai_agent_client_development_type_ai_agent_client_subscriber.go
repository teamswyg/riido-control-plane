package riidoaiserver

type aiAgentClientSubscriber struct {
	dashboardAgentID      string
	principal             AuthorizationResult
	visibilityKey         string
	events                chan ClientStreamEvent
	droppedEvents         int64
	terminalCompensations int64
}
