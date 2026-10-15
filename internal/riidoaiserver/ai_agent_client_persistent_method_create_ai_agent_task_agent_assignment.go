package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) CreateAIAgentTaskAgentAssignment(ctx context.Context, principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	return s.mutateTaskAgentMembership(ctx, taskID, req.AgentID, req.membershipPublicationGeneration, nil, true, func() (AIAgentTaskActionResponse, error) {
		return s.DevelopmentAIAgentClientStore.CreateAIAgentTaskAgentAssignment(ctx, principal, taskID, req)
	})
}
