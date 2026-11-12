package riidoaiserver

import "strings"

type composedAssignRequest struct {
	Request               AssignRequest
	IntentGateRequired    bool
	AgentOwnerPrincipalID string
	// TaskContext is captured for CLI reads; native follow-ups omit it from the prompt.
	TaskContext *AIAgentTaskContext
}

func composeAssignRequestWithTaskContext(taskID, componentID string, req AssignRequest, contextSnapshot AIAgentTaskContext) (AssignRequest, error) {
	composed, err := composeAssignRequestWithTaskContextResult(taskID, componentID, req, contextSnapshot)
	if err != nil {
		return AssignRequest{}, err
	}
	return composed.Request, nil
}

func composeAssignRequestWithTaskContextResult(
	taskID, componentID string,
	req AssignRequest,
	contextSnapshot AIAgentTaskContext,
) (composedAssignRequest, error) {
	composed, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{
		TaskID:          taskID,
		UserInstruction: req.Prompt,
		Context:         contextSnapshot,
	})
	if err != nil {
		return composedAssignRequest{}, err
	}
	instruction := req.Prompt
	req.Prompt = composed.Prompt
	req.TeamID = contextSnapshot.TeamID
	if composed.HasRepository {
		req.Worktree = assignmentWorktreeFromTaskContext(contextSnapshot, composed.SelectedRepository)
	}
	if err := applyTaskCheckoutInstruction(req.Worktree, contextSnapshot.linkedPullRequests, instruction); err != nil {
		return composedAssignRequest{}, err
	}
	if strings.TrimSpace(req.ComponentID) == "" {
		req.ComponentID = strings.TrimSpace(contextSnapshot.Component.ID)
		if req.ComponentID == "" {
			req.ComponentID = componentID
		}
	}
	return composedAssignRequest{
		Request:            req,
		IntentGateRequired: composed.IntentGateRequired,
		TaskContext:        &contextSnapshot,
	}, nil
}

func composeAssignRequestWithoutTaskContext(taskID, componentID string, req AssignRequest) (AssignRequest, error) {
	if err := applyTaskCheckoutInstruction(nil, nil, req.Prompt); err != nil {
		return AssignRequest{}, err
	}
	composed, err := composeAIAgentAssignmentFallback(taskID, componentID, req.Prompt)
	if err != nil {
		return AssignRequest{}, err
	}
	req.Prompt = composed.Prompt
	if strings.TrimSpace(req.ComponentID) == "" {
		req.ComponentID = strings.TrimSpace(componentID)
	}
	return req, nil
}
