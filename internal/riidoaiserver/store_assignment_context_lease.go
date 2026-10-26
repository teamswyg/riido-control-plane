package riidoaiserver

import "context"

// assignmentLeaseChecker reports whether an assignment's lease is still held,
// the same expiry rule claim/heartbeat reconciliation uses.
type assignmentLeaseChecker interface {
	AssignmentLeaseCurrent(ctx context.Context, assignment Assignment) (bool, error)
}

// AssignmentLeaseCurrent lets read-only callers refuse an expired lease before
// reconciliation fails the assignment. A recent update proves the lease; else
// the durable active-assignment lease decides. Without one (in-memory store)
// the projection state is authoritative.
func (s *Store) AssignmentLeaseCurrent(ctx context.Context, assignment Assignment) (bool, error) {
	now := s.now()
	if !assignment.UpdatedAt.IsZero() && assignment.UpdatedAt.Add(s.activeLeaseDuration).After(now) {
		return true, nil
	}
	leaseStore, ok := s.operationStore.(AssignmentActiveLeaseStore)
	if !ok {
		return true, nil
	}
	lease, exists, err := leaseStore.LoadAgentActiveAssignment(ctx, assignment.AgentID)
	if err != nil {
		return false, err
	}
	return exists && lease.ActiveAssignmentID == assignment.ID && !lease.Expired(now), nil
}
