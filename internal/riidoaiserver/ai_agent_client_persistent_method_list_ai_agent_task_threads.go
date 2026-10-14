package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) ListAIAgentTaskThreads(ctx context.Context, principal AuthorizationResult, taskID string) (AIAgentTaskThreadCollectionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskThreadCollectionResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentTaskThreads(ctx, principal, taskID)
}
