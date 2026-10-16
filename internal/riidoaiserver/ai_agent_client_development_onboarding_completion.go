package riidoaiserver

import "context"

// CompleteAIAgentOnboardingTask adds the onboarding agent as a task participant and
// records one completed, agent-authored task_completed comment. It never calls the
// assignment queue and never starts execution. Eligibility rules are in
// ai_agent_onboarding_completion_guard.go.
func (s *DevelopmentAIAgentClientStore) CompleteAIAgentOnboardingTask(ctx context.Context, principal AuthorizationResult, taskID string, req CompleteAIAgentOnboardingTaskRequest) (AIAgentTaskActionResponse, error) {
	return s.completeAIAgentOnboardingTask(ctx, principal, taskID, req, true)
}
