package riidoaiserver

import (
	"context"
	"time"
)

func (s *Store) RecordAgentEvent(ctx context.Context, agentID string, req AgentEventRequest) (response AgentEventResponse, err error) {
	ctx, span := s.startStoreOperationTrace(ctx, StoreOperationAppendEvent)
	startedAt := time.Now()
	defer func() {
		s.observeStoreOperation(StoreOperationAppendEvent, startedAt, err)
		FinishTraceSpan(span, err)
	}()
	reply := make(chan eventResult, 1)
	if err := s.send(ctx, eventCmd{agentID: agentID, req: req, reply: reply}); err != nil {
		return AgentEventResponse{}, err
	}
	select {
	case res := <-reply:
		return res.response, res.err
	case <-ctx.Done():
		return AgentEventResponse{}, ctx.Err()
	}
}
