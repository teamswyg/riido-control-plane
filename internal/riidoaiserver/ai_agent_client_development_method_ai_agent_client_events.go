package riidoaiserver

import (
	"context"
)

func (s *DevelopmentAIAgentClientStore) AIAgentClientEvents(ctx context.Context, principal AuthorizationResult) ([]ClientStreamEvent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.clientEventsForPrincipalLocked(principal), nil
}
