package riidoaiserver

import (
	"context"
	"net/http"
)

// Runs in one subscriber's HTTP goroutine, never in shared event fanout or
// under the store lock. No positive cache survives delivery of an event.
func (s Server) clientContextReadForDelivery(r *http.Request, principal AuthorizationResult, event ClientStreamEvent) (ClientStreamEvent, bool) {
	read, ok := event.Payload.(AgentContextReadEvent)
	filter, supported := s.aiAgent.(aiAgentClientEventFilter)
	token, authenticated := requestToken(r)
	if r.Context().Err() != nil || !ok || !supported || !authenticated || s.config.Authorizer == nil || read.TaskID == "" || read.AgentID == "" || read.ThreadID == "" || read.RunID == "" || read.AssignmentID == "" {
		return ClientStreamEvent{}, false
	}
	ctx, cancel := context.WithTimeout(r.Context(), DefaultAIAgentTaskContextClientTimeout)
	defer cancel()
	current, err := s.config.Authorizer.Authorize(ctx, token, AuthorizationRequest{
		Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead,
		WorkspaceID: principal.WorkspaceID, TaskID: read.TaskID,
	})
	if err != nil || current.PrincipalID != principal.PrincipalID || (current.WorkspaceID != "" && current.WorkspaceID != principal.WorkspaceID) {
		return ClientStreamEvent{}, false
	}
	current.WorkspaceID = principal.WorkspaceID
	visible, err := s.loadCurrentRunReadSources(ctx, token, current.WorkspaceID, read.TaskID, read.AssignmentID)
	if err != nil {
		return ClientStreamEvent{}, false
	}
	event, allowed := filter.FilterAIAgentClientEvent(ctx, current, event)
	if !allowed || ctx.Err() != nil {
		return ClientStreamEvent{}, false
	}
	read = event.Payload.(AgentContextReadEvent)
	sources := make([]AgentContextReadSource, 0, len(read.Sources))
	for _, source := range read.Sources {
		if visible[runReadSourceKey{source.Kind, source.ID, source.Part}] || (source.Kind == "agent" && source.Part == "profile") {
			sources = append(sources, source)
		}
	}
	read.Sources = sources
	// Capped/replayed events do not retain every omitted source identity.
	// Never expose an unverifiable hidden count; GET reads supplies pagination.
	read.OmittedSources, read.omittedAgentIDs, read.omittedOther = 0, nil, 0
	event.Payload = read
	return event, len(sources) > 0
}
