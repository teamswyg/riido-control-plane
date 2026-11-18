package riidoaiserver

import (
	"context"
	"fmt"
	"time"
)

// UpdatedAt tracks assignment state changes, not durable lease heartbeats.
// Resolve every candidate before changing replay state or persisting repairs:
// a failed lease read must never be interpreted as an expired execution.
func replayExpiredActiveAssignmentIDs(ctx context.Context, state *storeState, candidates []string, now time.Time, leases AssignmentActiveLeaseStore) ([]string, error) {
	if leases == nil {
		return candidates, nil
	}
	expired := make([]string, 0, len(candidates))
	for _, id := range candidates {
		assignment := state.assignments[id]
		leaseCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		lease, exists, err := leases.LoadAgentActiveAssignment(leaseCtx, assignment.AgentID)
		cancel()
		if err != nil {
			return nil, fmt.Errorf("load active assignment lease during replay: %w", err)
		}
		if exists && lease.ActiveAssignmentID == assignment.ID && !lease.Expired(now) {
			continue
		}
		expired = append(expired, id)
	}
	return expired, nil
}
