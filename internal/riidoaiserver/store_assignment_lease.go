package riidoaiserver

import (
	"context"
	"strings"
	"time"
)

func (s *Store) assignmentActiveLeaseExpired(state *storeState, assignment Assignment, at time.Time) (bool, error) {
	if !assignmentHoldsActiveLease(assignment.State) {
		return false, nil
	}
	if !assignment.UpdatedAt.IsZero() && assignment.UpdatedAt.Add(s.activeLeaseDuration).After(at) {
		return false, nil
	}
	leaseStore, ok := s.operationStore.(AssignmentActiveLeaseStore)
	if !ok {
		return false, nil
	}
	lease, exists, err := leaseStore.LoadAgentActiveAssignment(context.Background(), assignment.AgentID)
	if err != nil {
		return false, err
	}
	if !exists {
		return true, nil
	}
	if lease.ActiveAssignmentID != assignment.ID {
		return true, nil
	}
	if lease.Expired(at) {
		return true, nil
	}
	if !lease.HeartbeatAt.IsZero() {
		assignment.UpdatedAt = lease.HeartbeatAt
		state.assignments[assignment.ID] = assignment
	}
	return false, nil
}

func (s *Store) failStaleAssignment(state *storeState, assignment Assignment) Assignment {
	return s.failStaleAssignmentWithMessage(state, assignment, "active assignment lease expired", nil)
}

func (s *Store) failStaleAssignmentWithMessage(state *storeState, assignment Assignment, message string, metadata map[string]string) Assignment {
	now := s.now()
	assignment.State = AssignmentFailed
	assignment.UpdatedAt = now
	state.assignments[assignment.ID] = assignment
	eventMetadata := map[string]string{"lease_token": assignment.LeaseToken}
	for key, value := range metadata {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key != "" && value != "" {
			eventMetadata[key] = value
		}
	}
	s.appendEvent(state, assignment.TaskID, assignment.ID, assignment.AgentID, EventAssignmentFailed, assignment.State, message, eventMetadata, now)
	return assignment
}
