package riidoaiserver

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDynamoDBAIAgentClientSnapshotWritesManifestAfterParts(t *testing.T) {
	fixedNow := fixedSnapshotTestNow()
	fixture := newSnapshotDynamoDBFixture(t, fixedNow, nil, nil)
	defer fixture.close()
	if err := fixture.store.SaveAIAgentClientSnapshot(context.Background(), snapshotTestRecord(fixedNow)); err != nil {
		t.Fatalf("SaveAIAgentClientSnapshot: %v", err)
	}
	requests := drainDynamoDBAIAgentClientSnapshotRequests(fixture.requests)
	if len(requests) != 1 {
		t.Fatalf("write requests = %d, want %d", len(requests), len(dynamoDBAIAgentClientSnapshotPartNames))
	}
	last := requests[len(requests)-1]
	assertDynamoDBTarget(t, last, dynamoDBTransactWriteTarget)
	var payload struct {
		TransactItems []struct {
			Put struct{ Item map[string]map[string]string }
		}
	}
	if err := json.Unmarshal(last.body, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.TransactItems) != len(dynamoDBAIAgentClientSnapshotPartNames)+1 {
		t.Fatalf("commit actions = %d", len(payload.TransactItems))
	}

	assertDynamoDBString(t, payload.TransactItems[len(payload.TransactItems)-1].Put.Item, "sk", dynamoDBAIAgentClientSnapshotSK)
}

func TestDynamoDBAIAgentClientSnapshotWritePressure(t *testing.T) {
	items := []map[string]map[string]string{
		{"sk": {"S": dynamoDBAIAgentClientSnapshotPartSK(dynamoDBAIAgentClientSnapshotPartAgents)}},
		{"sk": {"S": dynamoDBAIAgentClientSnapshotSK}},
	}
	stats := dynamoDBAIAgentClientSnapshotWritePressure(items)
	if stats.itemsWritten != 2 || stats.partsWritten != 1 || stats.partsSkipped != len(dynamoDBAIAgentClientSnapshotPartNames)-1 {
		t.Fatalf("stats = %+v", stats)
	}
}

func drainDynamoDBAIAgentClientSnapshotRequests(requests <-chan capturedDynamoDBRequest) []capturedDynamoDBRequest {
	out := make([]capturedDynamoDBRequest, 0, len(requests))
	for len(requests) > 0 {
		out = append(out, <-requests)
	}
	return out
}
