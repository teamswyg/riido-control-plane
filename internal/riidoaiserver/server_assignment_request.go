package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s Server) assignRequestFromAIAgentClientTask(ctx context.Context, principal AuthorizationResult, bearerToken, taskID string, req AssignAIAgentTaskRequest) (AssignRequest, error) {
	taskID = strings.TrimSpace(taskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	if taskID == "" {
		return AssignRequest{}, errors.New("task_id is required")
	}
	if req.AgentID == "" {
		return AssignRequest{}, errors.New("agent_id is required")
	}
	agents, err := s.aiAgent.ListAIAgentTaskAssignableAgents(ctx, principal, taskID)
	if err != nil {
		return AssignRequest{}, err
	}
	selected, err := selectAssignableAgent(agents.Agents, req.AgentID)
	if err != nil {
		return AssignRequest{}, err
	}
	binding, runtime, err := s.resolveAgentRuntimeFact(selected.AgentID)
	if err != nil {
		return AssignRequest{}, err
	}
	assignmentReq := AssignRequest{
		ComponentID:              taskID,
		AgentID:                  selected.AgentID,
		RuntimeProvider:          binding.RuntimeProvider,
		ModelID:                  selected.ModelID,
		AgentInstruction:         augmentAgentInstruction(selected.Instruction),
		AllowExperimentalRuntime: runtime.RequiresExperimentalOptIn,
		CreatedBy:                strings.TrimSpace(principal.PrincipalID),
	}
	return s.assignRequestWithTaskContextPromptForClient(ctx, taskID, assignmentReq, AIAgentTaskContextRequest{
		ComponentID: taskID,
		WorkspaceID: principal.WorkspaceID,
		BearerToken: bearerToken,
	})
}

func selectAssignableAgent(agents []AgentClientRecord, agentID string) (AgentClientRecord, error) {
	for _, agent := range agents {
		if agent.AgentID == agentID {
			return agent, nil
		}
	}
	return AgentClientRecord{}, ErrAIAgentNotFound
}

func (s Server) resolveAgentRuntimeFact(agentID string) (AgentRuntimeBinding, RuntimeRecord, error) {
	if factRegistry, ok := s.aiAgent.(AgentRuntimeFactRegistry); ok {
		binding, runtime, found := factRegistry.LookupAgentRuntimeFact(agentID)
		if !found {
			return AgentRuntimeBinding{}, RuntimeRecord{}, errors.New("ai agent runtime binding is not configured")
		}
		return binding, runtime, nil
	}
	registry, ok := s.aiAgent.(AgentRegistry)
	if !ok {
		return AgentRuntimeBinding{}, RuntimeRecord{}, errors.New("ai agent runtime registry is not configured")
	}
	binding, found := registry.LookupAgent(agentID)
	if !found {
		return AgentRuntimeBinding{}, RuntimeRecord{}, errors.New("ai agent runtime binding is not configured")
	}
	return binding, RuntimeRecord{}, nil
}
