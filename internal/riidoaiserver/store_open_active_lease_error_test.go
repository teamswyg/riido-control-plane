package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestOpenStoreWithConfigDoesNotRepairWhenReplayLeaseReadFails(t *testing.T) {
	for _, wantErr := range []error{errors.New("lease store unavailable"), context.Canceled, context.DeadlineExceeded} {
		t.Run(wantErr.Error(), func(t *testing.T) {
			now := time.Date(2026, 11, 12, 4, 40, 34, 0, time.UTC)
			operations := &replayLeaseReadErrorStore{runtimeFakeActiveLeaseOperationStore: *replayActiveLeaseFixture(now, AssignmentRunning), err: wantErr}
			first := operations.records[0].Assignment
			second := first
			second.ID, second.AgentID = "asn-2", "agent-2"
			operations.records = append(operations.records, runtimeAssignmentOperationRecord(second, EventAssignmentRunning, 2))
			operations.projections[second.ID] = AssignmentProjection{Assignment: second, LastEventSeq: 2}
			store, err := OpenStoreWithConfig(context.Background(), StoreConfig{
				Now: func() time.Time { return now }, ActiveLeaseDuration: 5 * time.Minute, OperationStore: operations,
			})
			if store != nil {
				defer store.Close()
			}
			if store != nil || !errors.Is(err, wantErr) {
				t.Fatalf("store=%v err=%v want no store and %v", store, err, wantErr)
			}
			// The first candidate has no lease; the second read fails. Neither
			// candidate may be persisted as failed after an incomplete check.
			if len(operations.records) != 2 || operations.calls != 2 {
				t.Fatalf("records=%d calls=%d", len(operations.records), operations.calls)
			}
		})
	}
}

func TestOpenStoreWithConfigBoundsReplayLeaseReadByCallerContext(t *testing.T) {
	now := time.Date(2026, 11, 12, 4, 40, 34, 0, time.UTC)
	operations := &replayLeaseReadErrorStore{runtimeFakeActiveLeaseOperationStore: *replayActiveLeaseFixture(now, AssignmentRunning), wait: true}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store, err := OpenStoreWithConfig(ctx, StoreConfig{
		Now: func() time.Time { return now }, ActiveLeaseDuration: 5 * time.Minute, OperationStore: operations,
	})
	if store != nil {
		defer store.Close()
	}
	if store != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("store=%v err=%v want context cancellation", store, err)
	}
	if len(operations.records) != 1 || !operations.bounded {
		t.Fatalf("records=%d bounded=%v", len(operations.records), operations.bounded)
	}
}
