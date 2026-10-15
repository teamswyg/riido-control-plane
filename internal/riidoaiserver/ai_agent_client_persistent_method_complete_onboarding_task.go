package riidoaiserver

import (
	"context"
	"strings"
)

func (s *PersistentAIAgentClientStore) CompleteAIAgentOnboardingTask(ctx context.Context, principal AuthorizationResult, taskID string, req CompleteAIAgentOnboardingTaskRequest) (AIAgentTaskActionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	// The membership key must be the normalized ID, the same one the development
	// store uses, or a padded agent_id would be recorded under a different key.
	taskID = strings.TrimSpace(taskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	// Not a publication: the membership carries no durable generation, so removal
	// fences treat it as a plain, non-executing participant.
	return s.mutateTaskAgentMembership(ctx, taskID, req.AgentID, "", nil, false, func() (AIAgentTaskActionResponse, error) {
		return s.DevelopmentAIAgentClientStore.CompleteAIAgentOnboardingTask(ctx, principal, taskID, req)
	})
}
