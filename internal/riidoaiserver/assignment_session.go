package riidoaiserver

import "strings"

const assignmentMetadataProviderSessionID = "provider_session_id"

func providerSessionIDFromAgentEventRequest(req AgentEventRequest) string {
	if value := strings.TrimSpace(req.ProviderSessionID); value != "" {
		return value
	}
	return strings.TrimSpace(req.Metadata[assignmentMetadataProviderSessionID])
}

func providerSessionIDFromTaskEvent(event TaskEvent) string {
	return strings.TrimSpace(event.Metadata[assignmentMetadataProviderSessionID])
}

func providerSessionIDFromAssignmentEvent(req AgentEventRequest, event TaskEvent) string {
	if value := providerSessionIDFromAgentEventRequest(req); value != "" {
		return value
	}
	return providerSessionIDFromTaskEvent(event)
}

func metadataWithProviderSessionID(metadata map[string]string, providerSessionID string) map[string]string {
	providerSessionID = strings.TrimSpace(providerSessionID)
	if providerSessionID == "" {
		return cloneMetadata(metadata)
	}
	out := cloneMetadata(metadata)
	if out == nil {
		out = map[string]string{}
	}
	out[assignmentMetadataProviderSessionID] = providerSessionID
	return out
}

func enrichAgentEventRequestWithAssignment(req AgentEventRequest, assignment *Assignment) AgentEventRequest {
	if assignment == nil {
		return req
	}
	if strings.TrimSpace(req.RuntimeProvider) == "" {
		req.RuntimeProvider = assignment.RuntimeProvider
	}
	if strings.TrimSpace(req.ModelID) == "" {
		req.ModelID = assignment.ModelID
	}
	if strings.TrimSpace(req.ProviderSessionID) == "" {
		req.ProviderSessionID = assignment.ProviderSessionID
	}
	if providerSessionID := providerSessionIDFromAgentEventRequest(req); providerSessionID != "" {
		req.ProviderSessionID = providerSessionID
		req.Metadata = metadataWithProviderSessionID(req.Metadata, providerSessionID)
	}
	return req
}
