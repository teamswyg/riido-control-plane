package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) GetAIAgentTaskThreadStreamSubscription(ctx context.Context, principal AuthorizationResult, taskID string) (AIAgentTaskThreadStreamSubscriptionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskThreadStreamSubscriptionResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.GetAIAgentTaskThreadStreamSubscription(ctx, principal, taskID)
}
