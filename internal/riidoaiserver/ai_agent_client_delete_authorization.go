package riidoaiserver

import (
	"context"
	"strings"
)

// AuthorizeAIAgentDeletion checks mutation access before cancelling durable work.
// Public visibility permits reads and must not authorize deletion side effects.
func (s *DevelopmentAIAgentClientStore) AuthorizeAIAgentDeletion(ctx context.Context, principal AuthorizationResult, agentID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.agents[strings.TrimSpace(agentID)]
	if !ok || !s.aiAgentMutableBy(principal, agent) {
		return ErrAIAgentNotFound
	}
	return nil
}

func (s *PersistentAIAgentClientStore) AuthorizeAIAgentDeletion(ctx context.Context, principal AuthorizationResult, agentID string) error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.forceReloadMembershipSnapshot(ctx); err != nil {
		return err
	}
	return s.DevelopmentAIAgentClientStore.AuthorizeAIAgentDeletion(ctx, principal, agentID)
}
