package riidoaiserver

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// Exercise snapshot restoration with the production lease/projection adapter.
// A rolling restart must only read these records, never fail the live execution
// or delete the active lease owned by the still-running daemon.
func TestOpenStoreWithConfigPreservesDynamoDBLeaseAfterSnapshotRestart(t *testing.T) {
	now := time.Date(2026, 11, 12, 4, 40, 34, 0, time.UTC)
	fixture := replayActiveLeaseFixture(now, AssignmentRunning)
	record := fixture.records[0]
	projection := sampleAssignmentProjectionDynamoDBItem(t, record.Assignment, 1)
	heartbeat := record
	heartbeat.RecordedAt = now.Add(-3 * time.Minute)
	lease := (&DynamoDBAssignmentOperationStore{activeLeaseDuration: 5 * time.Minute}).agentActiveAssignmentDynamoDBItem(heartbeat)
	var leaseReads, writes atomic.Int32
	server := httptest.NewServer(replayLeaseHandler(t, projection, lease, &leaseReads, &writes))
	defer server.Close()
	operations := newAssignmentOperationStoreHarnessStore(t, dynamoDBAssignmentOperationStoreHarnessConfig{
		AccessKeyID: "AKID", TableName: "assignments", Now: now,
	}, server.URL)
	defer operations.Close()
	snapshot, err := NewFileStoreSnapshot(filepath.Join(t.TempDir(), "snapshot.json"))
	if err != nil {
		t.Fatal(err)
	}
	state, err := stateFromAssignmentOperations(fixture.records)
	if err != nil {
		t.Fatal(err)
	}
	if err := snapshot.SaveStoreSnapshot(context.Background(), snapshotFromState(&state, record.RecordedAt)); err != nil {
		t.Fatal(err)
	}
	store, err := OpenStoreWithConfig(context.Background(), StoreConfig{
		Now: func() time.Time { return now }, ActiveLeaseDuration: 5 * time.Minute,
		SnapshotStore: snapshot, OperationStore: operations,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	metrics, err := store.Metrics(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if metrics.AssignmentsByState[AssignmentRunning] != 1 || metrics.AssignmentsByState[AssignmentFailed] != 0 {
		t.Fatalf("states after restart=%v", metrics.AssignmentsByState)
	}
	if leaseReads.Load() != 1 || writes.Load() != 0 {
		t.Fatalf("lease reads=%d mutations=%d", leaseReads.Load(), writes.Load())
	}
}
