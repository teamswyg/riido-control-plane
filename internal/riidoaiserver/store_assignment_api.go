package riidoaiserver

import (
	"context"
	"time"
)

func (s *Store) AssignTask(ctx context.Context, taskID string, req AssignRequest) (assignment Assignment, err error) {
	ctx, span := s.startStoreOperationTrace(ctx, StoreOperationCreateAssignment)
	startedAt := time.Now()
	defer func() {
		s.observeStoreOperation(StoreOperationCreateAssignment, startedAt, err)
		FinishTraceSpan(span, err)
	}()
	reply := make(chan assignResult, 1)
	if err := s.send(ctx, assignCmd{taskID: taskID, req: req, reply: reply}); err != nil {
		return Assignment{}, err
	}
	select {
	case res := <-reply:
		return res.assignment, res.err
	case <-ctx.Done():
		return Assignment{}, ctx.Err()
	}
}

func (s *Store) AssignTaskAdditive(ctx context.Context, taskID string, req AssignRequest) (assignment Assignment, err error) {
	ctx, span := s.startStoreOperationTrace(ctx, StoreOperationCreateAssignment)
	startedAt := time.Now()
	defer func() {
		s.observeStoreOperation(StoreOperationCreateAssignment, startedAt, err)
		FinishTraceSpan(span, err)
	}()
	reply := make(chan assignResult, 1)
	cmd := assignCmd{taskID: taskID, req: req, allowConcurrentTaskAgents: true, reply: reply}
	if err := s.send(ctx, cmd); err != nil {
		return Assignment{}, err
	}
	select {
	case res := <-reply:
		return res.assignment, res.err
	case <-ctx.Done():
		return Assignment{}, ctx.Err()
	}
}
