package riidoaiserver

import (
	"context"
	"time"
)

func replayActiveLeaseFixture(now time.Time, assignmentState AssignmentState) *runtimeFakeActiveLeaseOperationStore {
	assignment := leaseExpiryAssignment(assignmentState, now.Add(-25*time.Minute))
	assignment.CreatedAt = assignment.UpdatedAt
	assignment.LeaseToken = "lease-current"
	return &runtimeFakeActiveLeaseOperationStore{
		runtimeFakeAssignmentOperationStore: runtimeFakeAssignmentOperationStore{
			records: []AssignmentOperationRecord{runtimeAssignmentOperationRecord(assignment, EventAssignmentRunning, 1)},
		},
		activeFound: true,
		activeLease: AssignmentActiveLease{
			AgentID: assignment.AgentID, ActiveAssignmentID: assignment.ID, LeaseToken: assignment.LeaseToken,
			HeartbeatAt: now.Add(-3 * time.Minute), LeaseExpiresAt: now.Add(2 * time.Minute),
		},
		projections: map[string]AssignmentProjection{
			assignment.ID: {Assignment: assignment, LastEventSeq: 1},
		},
	}
}

type replayLeaseReadErrorStore struct {
	runtimeFakeActiveLeaseOperationStore
	err     error
	calls   int
	wait    bool
	bounded bool
}

func (s *replayLeaseReadErrorStore) LoadAgentActiveAssignment(ctx context.Context, agentID string) (AssignmentActiveLease, bool, error) {
	s.calls++
	deadline, ok := ctx.Deadline()
	s.bounded = ok && time.Until(deadline) <= 2*time.Second
	if s.wait {
		<-ctx.Done()
		return AssignmentActiveLease{}, false, ctx.Err()
	}
	if agentID == "agent-1" {
		return AssignmentActiveLease{}, false, nil
	}
	return AssignmentActiveLease{}, false, s.err
}
