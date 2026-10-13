package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) DeleteAIAgentTaskAgentAssignment(ctx context.Context, principal AuthorizationResult, taskID, agentID string, req AgentAssignmentActionRequest) (AIAgentTaskActionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.DeleteAIAgentTaskAgentAssignment(ctx, principal, taskID, agentID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}
