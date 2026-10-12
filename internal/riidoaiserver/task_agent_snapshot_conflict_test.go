package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

func TestSnapshotConflictPreservesWinnerParts(t *testing.T) {
	f := newSnapshotDynamoDBFixture(t, fixedSnapshotTestNow(), nil, nil)
	defer f.close()
	winner := snapshotTestRecord(fixedSnapshotTestNow())
	winner.Version = 1
	winner.TaskAssignedAgents = map[string][]string{"task": {"winner"}}
	if err := f.store.SaveAIAgentClientSnapshot(context.Background(), winner); err != nil {
		t.Fatal(err)
	}
	loser := winner
	loser.TaskAssignedAgents = map[string][]string{"task": {"loser"}}
	if err := f.store.SaveAIAgentClientSnapshot(context.Background(), loser); !errors.Is(err, errAIAgentSnapshotConflict) {
		t.Fatalf("stale writer accepted: %v", err)
	}
	loaded, ok, err := f.store.LoadAIAgentClientSnapshot(context.Background())
	if err != nil || !ok || loaded.TaskAssignedAgents["task"][0] != "winner" {
		t.Fatalf("winner corrupted: %+v %v", loaded, err)
	}
}
func TestMixedSnapshotPartsAreRejected(t *testing.T) {
	record := snapshotTestRecord(fixedSnapshotTestNow())
	items, _, _, err := encodeSplitDynamoDBAIAgentClientSnapshot(record, nil)
	if err != nil {
		t.Fatal(err)
	}
	items[0]["part_gzip"] = map[string]string{"S": "broken"}
	if _, _, err := decodeSplitDynamoDBAIAgentClientSnapshot(items); !errors.Is(err, errAIAgentSnapshotConflict) {
		t.Fatalf("mixed snapshot accepted: %v", err)
	}
}
