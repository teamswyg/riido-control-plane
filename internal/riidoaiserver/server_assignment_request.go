package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s Server) assignRequestFromAIAgentClientTask(ctx context.Context, principal AuthorizationResult, bearerToken, taskID string, req AssignAIAgentTaskRequest) (AssignRequest, error) {
	composed, err := s.assignRequestFromAIAgentClientTaskResult(ctx, principal, bearerToken, taskID, req)
	if err != nil {
		return AssignRequest{}, err
	}
	return composed.Request, nil
}

func (s Server) assignRequestFromAIAgentClientTaskResult(
	ctx context.Context,
	principal AuthorizationResult,
	bearerToken, taskID string,
	req AssignAIAgentTaskRequest,
) (composedAssignRequest, error) {
	taskID = strings.TrimSpace(taskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	if taskID == "" {
		return composedAssignRequest{}, errors.New("task_id is required")
	}
	if req.AgentID == "" {
		return composedAssignRequest{}, errors.New("agent_id is required")
	}
	agents, err := s.aiAgent.ListAIAgentTaskAssignableAgents(ctx, principal, taskID)
	if err != nil {
		return composedAssignRequest{}, err
	}
	selected, err := selectAssignableAgent(agents.Agents, req.AgentID)
	if err != nil {
		return composedAssignRequest{}, err
	}
	binding, runtime, err := s.resolveAgentRuntimeFact(selected.AgentID)
	if err != nil {
		return composedAssignRequest{}, err
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
	return s.assignRequestWithTaskContextPromptForClientResult(ctx, taskID, assignmentReq, AIAgentTaskContextRequest{
		ComponentID: taskID,
		WorkspaceID: principal.WorkspaceID,
		BearerToken: bearerToken,
	})
}
