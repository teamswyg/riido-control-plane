package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) UnassignAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req UnassignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	return s.mutateTaskAgentMembership(ctx, taskID, req.AgentID, "", req.membershipRemovalFence, false, func() (AIAgentTaskActionResponse, error) {
		return s.DevelopmentAIAgentClientStore.UnassignAIAgentTask(ctx, principal, taskID, req)
	})
}
