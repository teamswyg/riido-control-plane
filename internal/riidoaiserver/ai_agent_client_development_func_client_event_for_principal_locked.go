package riidoaiserver

func clientEventForPrincipalLocked(s *DevelopmentAIAgentClientStore, principal AuthorizationResult, event ClientStreamEvent) (ClientStreamEvent, bool) {
	if deviceEvent, ok := event.Payload.(DeviceRuntimeSnapshotEvent); ok {
		device, ok := s.visibleDeviceRecordLocked(principal, deviceEvent.Device)
		if !ok {
			return ClientStreamEvent{}, false
		}
		deviceEvent.Device = device
		event.Payload = deviceEvent
		return event, true
	}
	if progressEvent, ok := event.Payload.(AgentThreadProgressEvent); ok {
		progressEvent.Lines = copyClientVisibleProgressLines(progressEvent.Lines)
		event.Payload = progressEvent
	}
	if !clientEventVisibleToLocked(s, principal, event) {
		return ClientStreamEvent{}, false
	}
	return event, true
}
