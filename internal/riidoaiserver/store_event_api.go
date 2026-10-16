package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"time"
)

func (s *Store) RecordAgentEvent(ctx context.Context, agentID string, req AgentEventRequest) (response AgentEventResponse, err error) {
	ctx, span := s.startStoreOperationTrace(ctx, StoreOperationAppendEvent)
	startedAt := time.Now()
	defer func() {
		s.observeStoreOperation(StoreOperationAppendEvent, startedAt, err)
		FinishTraceSpan(span, err)
	}()
	if strings.TrimSpace(agentID) == "" {
		return AgentEventResponse{}, errors.New("agent_id is required")
	}
	if strings.TrimSpace(req.AssignmentID) == "" {
		return AgentEventResponse{}, errors.New("assignment_id is required")
	}
	approval, err := s.preflightRecoveryCandidate(ctx, agentID, req)
	if err != nil {
		return AgentEventResponse{}, err
	}
	reply := make(chan eventResult, 1)
	if err := s.send(ctx, eventCmd{ctx: ctx, approval: approval, agentID: agentID, req: req, reply: reply}); err != nil {
		return AgentEventResponse{}, assignmentDependencyFailure(err)
	}
	select {
	case res := <-reply:
		return res.response, res.err
	case <-ctx.Done():
		return AgentEventResponse{}, ctx.Err()
	}
}
