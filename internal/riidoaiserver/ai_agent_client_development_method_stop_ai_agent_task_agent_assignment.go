package riidoaiserver

import (
	"context"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) StopAIAgentTaskAgentAssignment(ctx context.Context, principal AuthorizationResult, taskID, agentID string, req AgentAssignmentActionRequest) (AIAgentTaskActionResponse, error) {
	response, err := s.StopAIAgentTask(ctx, principal, taskID, StopAIAgentTaskRequest{
		AgentID:      agentID,
		AssignmentID: strings.TrimSpace(req.AssignmentID),
		Reason:       req.Reason,
		durableState: req.durableState,
	})
	if err != nil {
		return response, err
	}
	return response, nil
}
