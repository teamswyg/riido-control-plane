package riidoaiserver

func clientVisibleWorkStatusChangedEvent(event AgentWorkStatusChangedEvent) AgentWorkStatusChangedEvent {
	event.ResultMessage = clientVisibleTaskThreadText(event.ResultMessage)
	event.FailureDiagnostics = clientVisibleFailureDiagnostics(event.FailureDiagnostics)
	event.ExecutionContext = copyExecutionContext(event.ExecutionContext)
	return event
}
