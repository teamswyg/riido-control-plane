package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) CompleteAIAgentOnboardingTask(ctx context.Context, principal AuthorizationResult, taskID string, req CompleteAIAgentOnboardingTaskRequest) (AIAgentTaskActionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	// Not a publication: the membership carries no durable generation, so removal
	// fences treat it as a plain, non-executing participant.
	return s.mutateTaskAgentMembership(ctx, taskID, req.AgentID, "", nil, false, func() (AIAgentTaskActionResponse, error) {
		return s.DevelopmentAIAgentClientStore.CompleteAIAgentOnboardingTask(ctx, principal, taskID, req)
	})
}
