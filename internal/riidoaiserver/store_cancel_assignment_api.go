package riidoaiserver

import (
	"context"
	"time"
)

func (s *Store) CancelAssignment(ctx context.Context, taskID string, req CancelAssignmentRequest) (assignment Assignment, err error) {
	ctx, span := s.startStoreOperationTrace(ctx, StoreOperationCancelAssignment)
	startedAt := time.Now()
	defer func() {
		s.observeStoreOperation(StoreOperationCancelAssignment, startedAt, err)
		FinishTraceSpan(span, err)
	}()
	reply := make(chan cancelAssignmentResult, 1)
	if err := s.send(ctx, cancelAssignmentCmd{taskID: taskID, req: req, reply: reply}); err != nil {
		return Assignment{}, err
	}
	select {
	case res := <-reply:
		return res.assignment, res.err
	case <-ctx.Done():
		return Assignment{}, ctx.Err()
	}
}
