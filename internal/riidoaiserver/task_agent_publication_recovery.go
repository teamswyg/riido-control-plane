package riidoaiserver

import (
	"context"
	"time"
)

type taskAgentPublicationRecoveryResult struct {
	projection AssignmentProjection
	operation  *AssignmentOperationRecord
}

type taskAgentPublicationRecoveryStore interface {
	recoverFailedTaskAgentPublication(context.Context, Assignment, int64) (*taskAgentPublicationRecoveryResult, error)
}

type recoverTaskAgentPublicationCmd struct {
	ctx        context.Context
	assignment Assignment
	reply      chan error
}

func (s Server) recoverTaskAgentPublication(ctx context.Context, assignment Assignment, receipt *taskAgentCreationReceipt) {
	if !receipt.ownCreation(assignment) {
		return
	}
	owner, ok := s.assignment.(interface {
		recoverFailedTaskAgentPublication(context.Context, Assignment) error
	})
	if !ok {
		return
	}
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	_ = owner.recoverFailedTaskAgentPublication(cleanup, assignment)
}

func (s *Store) recoverFailedTaskAgentPublication(ctx context.Context, assignment Assignment) error {
	if _, ok := s.operationStore.(taskAgentPublicationRecoveryStore); !ok {
		return nil
	}
	reply := make(chan error, 1)
	if err := s.send(ctx, recoverTaskAgentPublicationCmd{ctx: ctx, assignment: assignment, reply: reply}); err != nil {
		return err
	}
	select {
	case err := <-reply:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Store) applyTaskAgentPublicationRecovery(state *storeState, ctx context.Context, assignment Assignment) error {
	result, err := s.operationStore.(taskAgentPublicationRecoveryStore).recoverFailedTaskAgentPublication(ctx, assignment, state.nextEventSeq+1)
	if err != nil || result == nil {
		return err
	}
	applyAssignmentProjectionToState(state, result.projection)
	if result.operation != nil {
		for _, event := range result.operation.Events {
			s.appendRecordedEvent(state, event)
		}
	}
	s.signalAgentWaiters(state, assignment.AgentID)
	return nil
}
