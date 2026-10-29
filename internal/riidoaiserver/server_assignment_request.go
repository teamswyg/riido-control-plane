package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s Server) assignRequestFromAIAgentClientTaskResult(
	ctx context.Context,
	principal AuthorizationResult,
	bearerToken, taskID string,
	req AssignAIAgentTaskRequest,
) (composedAssignRequest, error) {
	taskID = strings.TrimSpace(taskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	if err := req.DocumentSnapshot.validate(); err != nil {
		return composedAssignRequest{}, err
	}
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
	policy, err := s.authorizeRecoveryPolicy(ctx, principal, selected, binding, runtime.RequiresExperimentalOptIn, req.RecoveryPolicy)
	if err != nil {
		return composedAssignRequest{}, err
	}
	if selected.ReferenceFolders != nil && !selected.ReferenceFolders.Matches(binding.DeviceID, binding.RuntimeID) {
		return composedAssignRequest{}, errors.New("reference_folders no longer matches the agent runtime; select the folders again")
	}
	assignmentReq := AssignRequest{
		RecoveryPolicy: policy,

		RequestID: req.RequestID, WorkspaceID: principal.WorkspaceID, Prompt: strings.TrimSpace(req.Instruction),

		ComponentID:              taskID,
		AgentID:                  selected.AgentID,
		RuntimeProvider:          binding.RuntimeProvider,
		ModelID:                  selected.ModelID,
		AgentInstruction:         augmentAgentInstruction(selected.Instruction),
		ReferenceFolders:         selected.ReferenceFolders.Clone(),
		AllowExperimentalRuntime: runtime.RequiresExperimentalOptIn,
		CreatedBy:                strings.TrimSpace(principal.PrincipalID),
	}
	var composed composedAssignRequest
	if req.VerifiedTaskContext != nil {
		composed, err = composeAssignRequestWithTaskContextResult(taskID, taskID, assignmentReq, *req.VerifiedTaskContext)
	} else {
		composed, err = s.assignRequestWithTaskContextPromptForClientResult(ctx, taskID, assignmentReq, AIAgentTaskContextRequest{
			ComponentID:      taskID,
			WorkspaceID:      principal.WorkspaceID,
			BearerToken:      bearerToken,
			DocumentSnapshot: req.DocumentSnapshot,
			Extended:         s.config.AssignmentContext != nil,
		})
	}
	// Profile ownership is distinct from the connected physical runtime ownership.
	composed.AgentOwnerPrincipalID = selected.OwnerPrincipalID
	return composed, err
}
