package riidoaiserver

import (
	"context"
	"time"
)

func (s *DevelopmentAIAgentClientStore) UpdateAIAgentConfiguration(ctx context.Context, principal AuthorizationResult, agentID string, req UpdateAgentConfigurationRequest) (AgentClientRecordResponse, error) {
	if err := ctx.Err(); err != nil {
		return AgentClientRecordResponse{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.agentForMutation(principal, agentID)
	if !ok {
		return AgentClientRecordResponse{}, ErrAIAgentNotFound
	}
	if agent.AssignedTaskCount > 0 {
		return AgentClientRecordResponse{}, ErrAIAgentAssigned
	}
	previousRuntimeID := agent.RuntimeID
	if err := applyAgentConfigurationPatch(&agent, req); err != nil {
		return AgentClientRecordResponse{}, err
	}
	if err := s.applyAgentRuntimePatchLocked(&agent, principal, previousRuntimeID, req); err != nil {
		return AgentClientRecordResponse{}, err
	}
	agent.Editability = editabilityForAssignedTasks(agent.AssignedTaskCount)
	agent.UpdatedAt = time.Now().UTC()
	s.agents[agent.AgentID] = agent
	return AgentClientRecordResponse{SchemaVersion: SchemaVersion, Agent: s.agentForPrincipal(agent, principal)}, nil
}
