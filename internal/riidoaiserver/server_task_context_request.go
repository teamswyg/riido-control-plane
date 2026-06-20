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

func (s Server) assignRequestWithTaskContextPromptForClient(ctx context.Context, taskID string, req AssignRequest, contextReq AIAgentTaskContextRequest) (AssignRequest, error) {
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
		return composeAssignRequestWithoutTaskContext(taskID, componentID, req)
	}
	composedReq, err := composeAssignRequestWithTaskContext(taskID, componentID, req, contextSnapshot)
	if err != nil {
		return composeAssignRequestWithoutTaskContext(taskID, componentID, req)
	}
	return composedReq, nil
}

func (s Server) getAIAgentTaskContextWithTrace(ctx context.Context, componentID string) (snapshot AIAgentTaskContext, err error) {
	ctx, span := startTaskContextOperationTrace(ctx, TaskContextOperationResolve)
	defer func() {
		FinishTraceSpan(span, err)
	}()
	return s.taskContext.GetAIAgentTaskContext(ctx, componentID)
}

func (s Server) getAIAgentTaskContextForRequestWithTrace(ctx context.Context, req AIAgentTaskContextRequest) (snapshot AIAgentTaskContext, err error) {
	ctx, span := startTaskContextOperationTrace(ctx, TaskContextOperationResolve)
	defer func() {
		FinishTraceSpan(span, err)
	}()
	return s.getAIAgentTaskContextForRequest(ctx, req)
}

func (s Server) getAIAgentTaskContextForRequest(ctx context.Context, req AIAgentTaskContextRequest) (AIAgentTaskContext, error) {
	if s.taskContext == nil {
		return AIAgentTaskContext{}, errors.New("task context reader is not configured")
	}
	if reader, ok := s.taskContext.(AIAgentTaskContextRequestReader); ok {
		return reader.GetAIAgentTaskContextForRequest(ctx, req)
	}
	return s.taskContext.GetAIAgentTaskContext(ctx, req.ComponentID)
}
