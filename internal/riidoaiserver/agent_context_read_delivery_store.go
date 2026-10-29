package riidoaiserver

import "context"

type aiAgentClientEventFilter interface {
	FilterAIAgentClientEvent(context.Context, AuthorizationResult, ClientStreamEvent) (ClientStreamEvent, bool)
}

// Delivery rechecks the existing agent/thread policy using the freshly
// authorized principal, including deleted-agent thread snapshots.
func (s *DevelopmentAIAgentClientStore) FilterAIAgentClientEvent(ctx context.Context, principal AuthorizationResult, event ClientStreamEvent) (ClientStreamEvent, bool) {
	ctx, cancel := context.WithTimeout(ctx, DefaultAIAgentTaskContextClientTimeout)
	defer cancel()
	if !lockContextReadDelivery(ctx, &s.mu) {
		return ClientStreamEvent{}, false
	}
	defer s.mu.Unlock()
	return clientEventForPrincipalLocked(s, principal, event)
}

func (s *PersistentAIAgentClientStore) FilterAIAgentClientEvent(ctx context.Context, principal AuthorizationResult, event ClientStreamEvent) (ClientStreamEvent, bool) {
	ctx, cancel := context.WithTimeout(ctx, DefaultAIAgentTaskContextClientTimeout)
	defer cancel()
	if !lockContextReadDelivery(ctx, &s.operationMu) {
		return ClientStreamEvent{}, false
	}
	defer s.operationMu.Unlock()
	if s.reloadSnapshot(ctx) != nil {
		return ClientStreamEvent{}, false
	}
	return s.DevelopmentAIAgentClientStore.FilterAIAgentClientEvent(ctx, principal, event)
}
