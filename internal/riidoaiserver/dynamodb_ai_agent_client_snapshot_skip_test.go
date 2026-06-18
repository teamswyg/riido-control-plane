package riidoaiserver

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestDynamoDBAIAgentClientSnapshotSkipsUnchangedSplitParts(t *testing.T) {
	fixedNow := fixedSnapshotTestNow()
	fixture := newSnapshotDynamoDBFixture(t, fixedNow, nil, nil)
	defer fixture.close()
	snapshot := snapshotTestRecord(fixedNow)
	if err := fixture.store.SaveAIAgentClientSnapshot(context.Background(), snapshot); err != nil {
		t.Fatalf("initial SaveAIAgentClientSnapshot: %v", err)
	}
	drainSnapshotRequests(fixture.requests)

	snapshot.NextDaemonCommand = 3
	snapshot.SavedAt = fixedNow.Add(time.Minute)
	if err := fixture.store.SaveAIAgentClientSnapshot(context.Background(), snapshot); err != nil {
		t.Fatalf("manifest-only SaveAIAgentClientSnapshot: %v", err)
	}
	if len(fixture.requests) != 1 {
		t.Fatalf("manifest-only save wrote %d items, want 1", len(fixture.requests))
	}
	var payload struct {
		Item map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal((<-fixture.requests).body, &payload); err != nil {
		t.Fatalf("decode manifest-only PutItem payload: %v", err)
	}
	assertDynamoDBString(t, payload.Item, "sk", dynamoDBAIAgentClientSnapshotSK)
	assertDynamoDBNumber(t, payload.Item, "next_daemon_command", "3")
	if len(fixture.items) != len(dynamoDBAIAgentClientSnapshotPartNames)+1 {
		t.Fatalf("unexpected stored item count = %d", len(fixture.items))
	}
}
