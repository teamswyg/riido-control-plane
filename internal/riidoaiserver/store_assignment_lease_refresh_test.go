package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestAssignmentActiveLeaseExpiredRefreshesHeartbeat(t *testing.T) {
	now := time.Date(2026, 7, 8, 2, 32, 0, 0, time.UTC)
	heartbeat := now.Add(-5 * time.Second)
	assignment := leaseExpiryAssignment(AssignmentRunning, now.Add(-2*time.Minute))
	ops := &runtimeFakeActiveLeaseOperationStore{
		activeFound: true,
		activeLease: AssignmentActiveLease{
			AgentID:            assignment.AgentID,
			ActiveAssignmentID: assignment.ID,
			HeartbeatAt:        heartbeat,
			LeaseExpiresAt:     now.Add(time.Minute),
		},
	}
	state := newStoreState()
	state.assignments[assignment.ID] = assignment

	expired, err := (&Store{activeLeaseDuration: time.Minute, operationStore: ops}).
		assignmentActiveLeaseExpired(&state, assignment, now)
	if err != nil || expired {
		t.Fatalf("expired=%v err=%v", expired, err)
	}
	if got := state.assignments[assignment.ID].UpdatedAt; !got.Equal(heartbeat) {
		t.Fatalf("updatedAt=%s want %s", got, heartbeat)
	}
}

func TestAssignmentActiveLeaseExpiredReturnsLeaseStoreError(t *testing.T) {
	now := time.Date(2026, 7, 8, 2, 33, 0, 0, time.UTC)
	wantErr := errors.New("lease load failed")
	assignment := leaseExpiryAssignment(AssignmentRunning, now.Add(-2*time.Minute))
	store := &Store{
		activeLeaseDuration: time.Minute,
		operationStore:      &errorActiveLeaseStore{err: wantErr},
	}
	state := newStoreState()
	expired, err := store.assignmentActiveLeaseExpired(&state, assignment, now)
	if expired || !errors.Is(err, wantErr) {
		t.Fatalf("expired=%v err=%v", expired, err)
	}
}

type errorActiveLeaseStore struct {
	runtimeFakeAssignmentOperationStore
	err error
}

func (s errorActiveLeaseStore) LoadAgentActiveAssignment(context.Context, string) (AssignmentActiveLease, bool, error) {
	return AssignmentActiveLease{}, false, s.err
}

func (s errorActiveLeaseStore) RefreshAgentActiveAssignment(context.Context, Assignment, time.Time) error {
	return nil
}

func leaseExpiryAssignment(state AssignmentState, updatedAt time.Time) Assignment {
	return Assignment{ID: "asn-1", TaskID: "task-1", AgentID: "agent-1", State: state, UpdatedAt: updatedAt}
}
