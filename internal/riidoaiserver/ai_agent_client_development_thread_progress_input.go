package riidoaiserver

type threadProgressInput struct {
	AgentID           string
	Request           AgentThreadProgressBatchRequest
	Lines             []AgentThreadProgressLine
	GeneratedThreadID bool
}
