package riidoaiserver

import (
	"testing"
	"time"
)

func TestCoverageAuthorityRejectsIndependentAdapters(t *testing.T) {
	snapshot := &DynamoDBAIAgentClientSnapshot{region: "fixture", tableName: "same-table", endpoint: "http://127.0.0.1:1"}
	operations := &DynamoDBAssignmentOperationStore{region: snapshot.region, tableName: snapshot.tableName, endpoint: snapshot.endpoint}
	if ExecutionCoverageAuthorityCompatible(snapshot, operations) {
		t.Fatal("matching endpoint strings proved unrelated credential authorities")
	}
}

func TestCoverageAuthorityFactoryUsesTheSnapshotAuthority(t *testing.T) {
	fixture := newMembershipSharedDynamoFixture(t)
	snapshot := newSnapshotDynamoDBStore(t, time.Now(), fixture.server.URL, nil)
	t.Cleanup(func() { _ = snapshot.Close() })
	operations, err := NewExecutionCoverageAssignmentStore(snapshot, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = operations.Close() })
	if !ExecutionCoverageAuthorityCompatible(snapshot, operations) {
		t.Fatal("factory failed to bind identical authority")
	}
	unrelated := newSnapshotDynamoDBStore(t, time.Now(), fixture.server.URL, nil)
	t.Cleanup(func() { _ = unrelated.Close() })
	if ExecutionCoverageAuthorityCompatible(unrelated, operations) {
		t.Fatal("separate provider with matching selectors was admitted")
	}
}
