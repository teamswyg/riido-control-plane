package riidoaiserver

func eventAgentID(payload any) (string, bool) {
	switch event := payload.(type) {
	case AgentEditabilityChangedEvent:
		return event.AgentID, true
	case AgentWorkStatusChangedEvent:
		return event.AgentID, true
	case AgentThreadProgressEvent:
		return event.AgentID, true
	case AgentContextReadEvent:
		return event.AgentID, true
	default:
		return "", false
	}
}
