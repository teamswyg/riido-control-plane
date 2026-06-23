package riidoaiserver

func eventTaskThreadRef(payload any) (string, string, bool) {
	switch event := payload.(type) {
	case AgentWorkStatusChangedEvent:
		return event.TaskID, event.ThreadID, event.TaskID != "" && event.ThreadID != ""
	case AgentThreadProgressEvent:
		return event.TaskID, event.ThreadID, event.TaskID != "" && event.ThreadID != ""
	default:
		return "", "", false
	}
}
