package riidoaiserver

import (
	"context"
	"maps"
)

type executionTerminationCommand struct {
	ctx      context.Context
	identity executionTerminationIdentity
	request  executionTerminationRequest
	reply    chan executionTerminationResult
}
type executionTerminationResult struct {
	receipt executionTerminationReceipt
	err     error
}

func (s *Store) recordExecutionTermination(ctx context.Context, id executionTerminationIdentity, req executionTerminationRequest) (executionTerminationReceipt, error) {
	if durable, ok := s.operationStore.(executionTerminationRecorder); ok {
		return durable.recordExecutionTermination(ctx, id, req)
	}
	if s.operationStore != nil {
		return executionTerminationReceipt{}, errTerminationUnavailable
	}
	reply := make(chan executionTerminationResult, 1)
	if err := s.send(ctx, executionTerminationCommand{ctx, id, req, reply}); err != nil {
		return executionTerminationReceipt{}, err
	}
	select {
	case result := <-reply:
		return result.receipt, result.err
	case <-ctx.Done():
		return executionTerminationReceipt{}, ctx.Err()
	}
}

func (s *Store) commitExecutionTermination(state *storeState, msg executionTerminationCommand) executionTerminationResult {
	if err := msg.ctx.Err(); err != nil {
		return executionTerminationResult{err: err}
	}
	claim, ok := state.executionClaims[msg.identity.key()]
	if !ok || !msg.identity.matches(claim) {
		return executionTerminationResult{err: errTerminationForbidden}
	}
	if err := executionTerminationProjectionGuard(msg.identity, state.assignments[msg.identity.assignmentID]); err != nil {
		return executionTerminationResult{err: err}
	}
	if old, ok := state.executionTerminations[msg.identity.key()]; ok {
		receipt, err := matchingExecutionTermination(old, msg.request)
		return executionTerminationResult{receipt, err}
	}
	receipt := newExecutionTermination(msg.identity, msg.request, s.now())
	trial := *state
	trial.devicePolicyRevisions = maps.Clone(state.devicePolicyRevisions)
	if err := bumpDevicePolicyRevision(&trial, msg.identity.deviceID); err != nil {
		return executionTerminationResult{err: err}
	}
	trial.executionTerminations = maps.Clone(state.executionTerminations)
	trial.executionTerminations[msg.identity.key()] = receipt
	if s.snapshotStore != nil {
		if err := s.snapshotStore.SaveStoreSnapshot(msg.ctx, snapshotFromState(&trial, s.now())); err != nil {
			return executionTerminationResult{err: err}
		}
	}
	state.executionTerminations = trial.executionTerminations
	state.devicePolicyRevisions = trial.devicePolicyRevisions
	return executionTerminationResult{receipt: receipt}
}
