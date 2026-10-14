package riidoaiserver

import (
	"context"
	"errors"
)

func (s Server) getAIAgentTaskContextWithTrace(
	ctx context.Context,
	componentID string,
) (snapshot AIAgentTaskContext, err error) {
	ctx, span := startTaskContextOperationTrace(ctx, TaskContextOperationResolve)
	defer func() {
		FinishTraceSpan(span, err)
	}()
	return s.taskContext.GetAIAgentTaskContext(ctx, componentID)
}

func (s Server) getAIAgentTaskContextForRequestWithTrace(
	ctx context.Context,
	req AIAgentTaskContextRequest,
) (snapshot AIAgentTaskContext, err error) {
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
