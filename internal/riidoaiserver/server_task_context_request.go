package riidoaiserver

import (
	"context"
	"strings"
)

func (s Server) assignRequestWithTaskContextPrompt(ctx context.Context, taskID string, req AssignRequest) (AssignRequest, error) {
	componentID := strings.TrimSpace(req.ComponentID)
	if componentID == "" {
		componentID = strings.TrimSpace(taskID)
	}
	contextSnapshot, err := s.getAIAgentTaskContextWithTrace(ctx, componentID)
	if err != nil {
		return AssignRequest{}, err
	}
	return composeAssignRequestWithTaskContext(taskID, componentID, req, contextSnapshot)
}

func (s Server) assignRequestWithTaskContextPromptForClientResult(
	ctx context.Context,
	taskID string,
	req AssignRequest,
	contextReq AIAgentTaskContextRequest,
) (composedAssignRequest, error) {
	componentID := strings.TrimSpace(req.ComponentID)
	if componentID == "" {
		componentID = strings.TrimSpace(taskID)
	}
	contextReq.ComponentID = strings.TrimSpace(contextReq.ComponentID)
	if contextReq.ComponentID == "" {
		contextReq.ComponentID = componentID
	}
	contextSnapshot, err := s.getAIAgentTaskContextForRequestWithTrace(ctx, contextReq)
	if err != nil {
		request, composeErr := composeAssignRequestWithoutTaskContext(taskID, componentID, req)
		return composedAssignRequest{Request: request}, composeErr
	}
	composedReq, err := composeAssignRequestWithTaskContextResult(taskID, componentID, req, contextSnapshot)
	if err != nil {
		request, composeErr := composeAssignRequestWithoutTaskContext(taskID, componentID, req)
		return composedAssignRequest{Request: request}, composeErr
	}
	return composedReq, nil
}
