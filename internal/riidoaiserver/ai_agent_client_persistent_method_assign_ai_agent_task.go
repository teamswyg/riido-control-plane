package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) AssignAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	return s.mutateTaskAgentMembership(ctx, taskID, req.AgentID, req.membershipPublicationGeneration, nil, true, func() (AIAgentTaskActionResponse, error) {
		return s.DevelopmentAIAgentClientStore.AssignAIAgentTask(ctx, principal, taskID, req)
	})
}
