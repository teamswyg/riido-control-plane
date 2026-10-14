package riidoaiserver

import "context"

type assignableAgentsErrorStore struct {
	*DevelopmentAIAgentClientStore
	err          error
	reconcileErr error
}

func (s assignableAgentsErrorStore) ListAIAgentTaskAssignableAgents(
	ctx context.Context,
	principal AuthorizationResult,
	taskID string,
) (AgentClientListResponse, error) {
	if s.err != nil {
		return AgentClientListResponse{}, s.err
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentTaskAssignableAgents(ctx, principal, taskID)
}

func (s assignableAgentsErrorStore) ReconcileAIAgentActiveThreadProjections(
	context.Context,
	AuthorizationResult,
	string,
	AssignmentProjectionReader,
) (bool, error) {
	return false, s.reconcileErr
}

func assignableAgentsErrorTestPath() string {
	return "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-assignable-errors/assignable-agents"
}
