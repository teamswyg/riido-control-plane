package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) ListAIAgentTaskAssignableAgents(ctx context.Context, principal AuthorizationResult, taskID string) (AgentClientListResponse, error) {
	if err := ctx.Err(); err != nil {
		return AgentClientListResponse{}, err
	}
	if strings.TrimSpace(taskID) == "" {
		return AgentClientListResponse{}, errors.New("task_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return AgentClientListResponse{SchemaVersion: SchemaVersion, Agents: s.visibleAgents(principal)}, nil
}
