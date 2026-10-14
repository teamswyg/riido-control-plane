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
	if len(requests) != len(dynamoDBAIAgentClientSnapshotPartNames)+1 {
		t.Fatalf("put requests = %d, want %d", len(requests), len(dynamoDBAIAgentClientSnapshotPartNames)+1)
	}
	last := decodeDynamoDBAIAgentClientSnapshotPut(t, requests[len(requests)-1])
	assertDynamoDBString(t, last.Item, "sk", dynamoDBAIAgentClientSnapshotSK)
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

func decodeDynamoDBAIAgentClientSnapshotPut(t *testing.T, request capturedDynamoDBRequest) struct {
	Item map[string]map[string]string `json:"Item"`
} {
	t.Helper()
	var payload struct {
		Item map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal(request.body, &payload); err != nil {
		t.Fatalf("decode PutItem payload: %v", err)
	}
	return payload
}

func drainDynamoDBAIAgentClientSnapshotRequests(requests <-chan capturedDynamoDBRequest) []capturedDynamoDBRequest {
	out := make([]capturedDynamoDBRequest, 0, len(requests))
	for len(requests) > 0 {
		out = append(out, <-requests)
	}
	return out
}
