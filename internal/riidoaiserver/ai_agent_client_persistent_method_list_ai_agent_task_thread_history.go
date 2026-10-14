package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) ListAIAgentTaskThreadHistory(ctx context.Context, principal AuthorizationResult, taskID string) (AIAgentTaskThreadHistoryCollectionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskThreadHistoryCollectionResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentTaskThreadHistory(ctx, principal, taskID)
}
