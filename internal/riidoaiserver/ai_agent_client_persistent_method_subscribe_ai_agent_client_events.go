package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) SubscribeAIAgentClientEvents(ctx context.Context, principal AuthorizationResult) ([]ClientStreamEvent, <-chan ClientStreamEvent, func(), error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return nil, nil, nil, err
	}
	return s.DevelopmentAIAgentClientStore.SubscribeAIAgentClientEvents(ctx, principal)
}
