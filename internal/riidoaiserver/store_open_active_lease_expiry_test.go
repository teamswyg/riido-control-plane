package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestOpenStoreWithConfigUsesDurableReplayLeaseExpiry(t *testing.T) {
	cases := []struct {
		name   string
		change func(*runtimeFakeActiveLeaseOperationStore, time.Time)
		want   AssignmentState
	}{
		{"older heartbeat with unexpired lease", func(s *runtimeFakeActiveLeaseOperationStore, now time.Time) {
			s.activeLease.HeartbeatAt = now.Add(-10 * time.Minute)
		}, AssignmentRunning},
		{"unexpired milliseconds without heartbeat", func(s *runtimeFakeActiveLeaseOperationStore, now time.Time) {
			s.activeLease.HeartbeatAt = time.Time{}
			s.activeLease.LeaseExpiresAt = time.Time{}
			s.activeLease.LeaseExpiresUnixMS = now.Add(time.Minute).UnixMilli()
		}, AssignmentRunning},
		{"expired", func(s *runtimeFakeActiveLeaseOperationStore, now time.Time) {
			s.activeLease.LeaseExpiresAt = now.Add(-time.Second)
		}, AssignmentFailed},
		{"expiry boundary", func(s *runtimeFakeActiveLeaseOperationStore, now time.Time) { s.activeLease.LeaseExpiresAt = now }, AssignmentFailed},
		{"expired milliseconds override timestamp", func(s *runtimeFakeActiveLeaseOperationStore, now time.Time) {
			s.activeLease.LeaseExpiresUnixMS = now.UnixMilli()
		}, AssignmentFailed},
		{"missing expiry", func(s *runtimeFakeActiveLeaseOperationStore, _ time.Time) { s.activeLease.LeaseExpiresAt = time.Time{} }, AssignmentFailed},
		{"missing lease", func(s *runtimeFakeActiveLeaseOperationStore, _ time.Time) { s.activeFound = false }, AssignmentFailed},
		{"other assignment", func(s *runtimeFakeActiveLeaseOperationStore, _ time.Time) {
			s.activeLease.ActiveAssignmentID = "asn-other"
		}, AssignmentFailed},
		{"new lease epoch for same assignment", func(s *runtimeFakeActiveLeaseOperationStore, _ time.Time) { s.activeLease.LeaseToken = "lease-new" }, AssignmentRunning},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Date(2026, 11, 12, 4, 40, 34, 0, time.UTC)
			operations := replayActiveLeaseFixture(now, AssignmentRunning)
			tc.change(operations, now)
			store, err := OpenStoreWithConfig(context.Background(), StoreConfig{
				Now: func() time.Time { return now }, ActiveLeaseDuration: time.Minute, OperationStore: operations,
			})
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			metrics, err := store.Metrics(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if metrics.AssignmentsByState[tc.want] != 1 {
				t.Fatalf("states=%v want %s", metrics.AssignmentsByState, tc.want)
			}
			wantRecords := 1
			if tc.want == AssignmentFailed {
				wantRecords++
			}
			if len(operations.records) != wantRecords {
				t.Fatalf("records=%d want %d", len(operations.records), wantRecords)
			}
		})
	}
}
