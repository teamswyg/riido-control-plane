package riidoaiserver

func clientVisibleWorkStatusChangedEvent(event AgentWorkStatusChangedEvent) AgentWorkStatusChangedEvent {
	event.ResultMessage = clientVisibleTaskThreadText(event.ResultMessage)
	event.FailureDiagnostics = clientVisibleFailureDiagnostics(event.FailureDiagnostics)
	return event
}
