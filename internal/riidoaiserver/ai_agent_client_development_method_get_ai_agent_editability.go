package riidoaiserver

import (
	"context"
)

func (s *DevelopmentAIAgentClientStore) GetAIAgentEditability(ctx context.Context, principal AuthorizationResult, agentID string) (AgentEditabilityResponse, error) {
	if err := ctx.Err(); err != nil {
		return AgentEditabilityResponse{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.visibleAgent(principal, agentID)
	if !ok {
		return AgentEditabilityResponse{}, ErrAIAgentNotFound
	}
	response := AgentEditabilityResponse{
		SchemaVersion:     SchemaVersion,
		AgentID:           agent.AgentID,
		Editability:       agent.Editability,
		AssignedTaskCount: agent.AssignedTaskCount,
	}
	if agent.Editability == AgentEditabilityBlockedAssignedTasks {
		response.Reason = "assigned_task_count must be zero before editing"
	}
	return response, nil
}
