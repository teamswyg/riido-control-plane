package riidoaiserver

import (
	"context"
)

func (s *DevelopmentAIAgentClientStore) BootstrapAIAgentClient(ctx context.Context, principal AuthorizationResult, clientKind ClientKind) (ClientBootstrapResponse, error) {
	if err := ctx.Err(); err != nil {
		return ClientBootstrapResponse{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return ClientBootstrapResponse{
		SchemaVersion: SchemaVersion,
		ClientKind:    normalizeClientKind(clientKind),
		WorkspaceID:   s.workspaceScope(principal),
		Agents:        s.visibleAgents(principal),
		Devices:       s.visibleDevicesLocked(principal),
	}, nil
}
