package riidoaiserver

import (
	"context"
	"time"
)

func (s *Store) loadDurableActiveAssignment(state *storeState, agentID string, at time.Time) (Assignment, *Assignment, bool, error) {
	leaseStore, hasLeaseStore := s.operationStore.(AssignmentActiveLeaseStore)
	projectionReader, hasProjectionReader := s.operationStore.(AssignmentProjectionReader)
	if !hasLeaseStore || !hasProjectionReader {
		return Assignment{}, nil, false, nil
	}
	lease, exists, err := leaseStore.LoadAgentActiveAssignment(context.Background(), agentID)
	if err != nil {
		return Assignment{}, nil, false, err
	}
	if !exists || lease.ActiveAssignmentID == "" {
		return Assignment{}, nil, false, nil
	}
	projection, exists, err := projectionReader.LoadAssignmentProjection(context.Background(), lease.ActiveAssignmentID)
	if err != nil {
		return Assignment{}, nil, false, err
	}
	if !exists {
		return Assignment{}, nil, false, nil
	}
	assignment := projection.Assignment
	if assignment.AgentID != agentID {
		return Assignment{}, nil, false, nil
	}
	applyAssignmentProjectionToState(state, projection)
	if !assignmentHoldsActiveLease(assignment.State) {
		return assignment, nil, true, nil
	}
	if lease.Expired(at) {
		stale := s.failStaleAssignment(state, assignment)
		return Assignment{}, &stale, false, nil
	}
	if !lease.HeartbeatAt.IsZero() && lease.HeartbeatAt.After(assignment.UpdatedAt) {
		assignment.UpdatedAt = lease.HeartbeatAt
		state.assignments[assignment.ID] = assignment
	}
	return assignment, nil, true, nil
}
