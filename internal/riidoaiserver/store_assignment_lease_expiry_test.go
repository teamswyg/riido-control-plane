package riidoaiserver

import (
	"testing"
	"time"
)

func TestAssignmentActiveLeaseExpiredFastPaths(t *testing.T) {
	now := time.Date(2026, 7, 8, 2, 30, 0, 0, time.UTC)
	store := &Store{activeLeaseDuration: time.Minute, operationStore: &runtimeFakeActiveLeaseOperationStore{}}
	state := newStoreState()

	queued := leaseExpiryAssignment(AssignmentQueued, now.Add(-time.Hour))
	expired, err := store.assignmentActiveLeaseExpired(&state, queued, now)
	if err != nil || expired {
		t.Fatalf("queued expired=%v err=%v", expired, err)
	}
	if store.operationStore.(*runtimeFakeActiveLeaseOperationStore).loadCalls != 0 {
		t.Fatal("queued assignment should not load active lease")
	}

	recent := leaseExpiryAssignment(AssignmentRunning, now.Add(-30*time.Second))
	expired, err = store.assignmentActiveLeaseExpired(&state, recent, now)
	if err != nil || expired {
		t.Fatalf("recent running expired=%v err=%v", expired, err)
	}
	if store.operationStore.(*runtimeFakeActiveLeaseOperationStore).loadCalls != 0 {
		t.Fatal("recent assignment should not load active lease")
	}
}

func TestAssignmentActiveLeaseExpiredFromDurableLease(t *testing.T) {
	now := time.Date(2026, 7, 8, 2, 31, 0, 0, time.UTC)
	assignment := leaseExpiryAssignment(AssignmentRunning, now.Add(-2*time.Minute))
	cases := []struct {
		name  string
		ops   AssignmentOperationStore
		want  bool
		calls int
	}{
		{name: "no active lease store", ops: &runtimeFakeAssignmentOperationStore{}, want: false},
		{name: "active lease missing", ops: &runtimeFakeActiveLeaseOperationStore{}, want: true, calls: 1},
		{name: "different active assignment", ops: &runtimeFakeActiveLeaseOperationStore{
			activeFound: true,
			activeLease: AssignmentActiveLease{
				AgentID:            assignment.AgentID,
				ActiveAssignmentID: "asn-other",
				LeaseExpiresAt:     now.Add(time.Minute),
			},
		}, want: true, calls: 1},
		{name: "expired active lease", ops: &runtimeFakeActiveLeaseOperationStore{
			activeFound: true,
			activeLease: AssignmentActiveLease{
				AgentID:            assignment.AgentID,
				ActiveAssignmentID: assignment.ID,
				LeaseExpiresAt:     now,
			},
		}, want: true, calls: 1},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{activeLeaseDuration: time.Minute, operationStore: tt.ops}
			state := newStoreState()
			expired, err := store.assignmentActiveLeaseExpired(&state, assignment, now)
			if err != nil || expired != tt.want {
				t.Fatalf("expired=%v err=%v", expired, err)
			}
			if ops, ok := tt.ops.(*runtimeFakeActiveLeaseOperationStore); ok && ops.loadCalls != tt.calls {
				t.Fatalf("loadCalls=%d want %d", ops.loadCalls, tt.calls)
			}
		})
	}
}
