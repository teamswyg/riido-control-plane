package riidoaiserver

type aiAgentClientSubscriber struct {
	principal AuthorizationResult
	events    chan ClientStreamEvent
}
