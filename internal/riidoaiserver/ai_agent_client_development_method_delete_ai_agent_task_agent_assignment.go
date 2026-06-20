package riidoaiserver

import (
	"context"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) DeleteAIAgentTaskAgentAssignment(ctx context.Context, principal AuthorizationResult, taskID, agentID string, req AgentAssignmentActionRequest) (AIAgentTaskActionResponse, error) {
	return s.UnassignAIAgentTask(ctx, principal, taskID, UnassignAIAgentTaskRequest{
		AgentID:      agentID,
		AssignmentID: strings.TrimSpace(req.AssignmentID),
		Reason:       req.Reason,
	})
}
