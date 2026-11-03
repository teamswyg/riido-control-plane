package riidoaiserver

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

func (s *Store) localAssignmentSignalVersions(state *storeState, ids []string) map[string]string {
	out := make(map[string]string, len(ids))
	for _, id := range ids {
		parts := []string{}
		for _, aid := range state.agentAssignments[id] {
			a := state.assignments[aid]
			if a.State.Code() == AssignmentStateCodeQueued || assignmentHoldsActiveLease(a.State) {
				parts = append(parts, assignmentSignalPart(a))
			}
		}
		sort.Strings(parts)
		out[id] = strings.Join(parts, "\n")
	}
	return out
}

func (s *Store) durableAssignmentSignalVersions(ctx context.Context, queue AssignmentQueueReader, ids []string) (map[string]string, error) {
	out := make(map[string]string, len(ids))
	for _, id := range ids {
		assignments, err := queue.LoadAgentQueueAssignments(ctx, id)
		if err != nil {
			return nil, err
		}
		parts := []string{}
		for _, a := range assignments {
			parts = append(parts, assignmentSignalPart(a))
		}
		leases, hasLeases := s.operationStore.(AssignmentActiveLeaseStore)
		projections, hasProjections := s.operationStore.(AssignmentProjectionReader)
		if hasLeases && hasProjections {
			lease, found, err := leases.LoadAgentActiveAssignment(ctx, id)
			if err != nil {
				return nil, err
			}
			if found && lease.ActiveAssignmentID != "" {
				projection, found, err := projections.LoadAssignmentProjection(ctx, lease.ActiveAssignmentID)
				if err != nil {
					return nil, err
				}
				if found && projection.Assignment.AgentID == id {
					parts = append(parts, fmt.Sprintf("%s:expired=%t", assignmentSignalPart(projection.Assignment), lease.Expired(s.now())))
				}
			}
		}
		sort.Strings(parts)
		out[id] = strings.Join(parts, "\n")
	}
	return out, nil
}

func assignmentSignalPart(a Assignment) string {
	// Exclude UpdatedAt, heartbeat and streaming text: ongoing output is not a new assignment.
	return fmt.Sprintf("%s:%s:%s:%s", a.ID, a.State, a.LeaseToken, a.BlockedByAssignmentID)
}
