package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) CreateAIAgentTaskAgentAssignment(ctx context.Context, principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.CreateAIAgentTaskAgentAssignment(ctx, principal, taskID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}
