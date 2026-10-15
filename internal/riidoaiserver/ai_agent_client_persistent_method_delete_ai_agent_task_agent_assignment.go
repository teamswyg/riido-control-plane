package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) DeleteAIAgentTaskAgentAssignment(ctx context.Context, principal AuthorizationResult, taskID, agentID string, req AgentAssignmentActionRequest) (AIAgentTaskActionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	return s.mutateTaskAgentMembership(ctx, taskID, agentID, "", req.membershipRemovalFence, false, func() (AIAgentTaskActionResponse, error) {
		return s.DevelopmentAIAgentClientStore.DeleteAIAgentTaskAgentAssignment(ctx, principal, taskID, agentID, req)
	})
}
