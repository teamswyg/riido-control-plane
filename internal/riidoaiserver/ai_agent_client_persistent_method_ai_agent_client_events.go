package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) AIAgentClientEvents(ctx context.Context, principal AuthorizationResult) ([]ClientStreamEvent, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return nil, err
	}
	return s.DevelopmentAIAgentClientStore.AIAgentClientEvents(ctx, principal)
}
