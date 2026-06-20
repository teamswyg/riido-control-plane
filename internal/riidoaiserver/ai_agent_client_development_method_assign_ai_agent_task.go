package riidoaiserver

import (
	"context"
)

func (s *DevelopmentAIAgentClientStore) AssignAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	return s.assignAIAgentTask(ctx, principal, taskID, req, true)
}
