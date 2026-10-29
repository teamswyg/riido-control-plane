package riidoaiserver

import (
	"context"
	"errors"
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
	if err := contextReq.DocumentSnapshot.validate(); err != nil {
		return composedAssignRequest{}, err
	}
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
		var latest latestTaskDocumentError
		if contextReq.DocumentSnapshot != nil {
			return composedAssignRequest{}, latestTaskDocumentError{err}
		}
		if errors.As(err, &latest) {
			return composedAssignRequest{}, err
		}
		request, composeErr := composeAssignRequestWithoutTaskContext(taskID, componentID, req)
		return composedAssignRequest{Request: request}, composeErr
	}
	if contextReq.DocumentSnapshot != nil {
		if contextSnapshot.Component.ID != contextReq.ComponentID {
			return composedAssignRequest{}, errors.New("task context component identity mismatch")
		}
		if contextSnapshot.ReadScope == nil {
			contextReq.DocumentSnapshot.apply(&contextSnapshot.Document)
		}
	}
	composedReq, err := composeAssignRequestWithTaskContextResult(taskID, componentID, req, contextSnapshot)
	if err != nil {
		if contextReq.DocumentSnapshot != nil {
			return composedAssignRequest{}, err
		}
		request, composeErr := composeAssignRequestWithoutTaskContext(taskID, componentID, req)
		return composedAssignRequest{Request: request}, composeErr
	}
	return composedReq, nil
}
