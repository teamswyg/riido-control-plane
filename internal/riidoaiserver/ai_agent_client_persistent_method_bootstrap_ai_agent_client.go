package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) BootstrapAIAgentClient(ctx context.Context, principal AuthorizationResult, clientKind ClientKind) (ClientBootstrapResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return ClientBootstrapResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.BootstrapAIAgentClient(ctx, principal, clientKind)
}
