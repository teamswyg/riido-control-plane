package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) SubmitAIAgentTaskComment(ctx context.Context, principal AuthorizationResult, taskID string, req SubmitAIAgentTaskCommentRequest) (AIAgentTaskActionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.SubmitAIAgentTaskComment(ctx, principal, taskID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}
