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
	winner.TaskRequestReceipts = map[string]AIAgentTaskActionResponse{"task": {AssignmentID: "winner"}}
	if err := f.store.SaveAIAgentClientSnapshot(context.Background(), winner); err != nil {
		t.Fatal(err)
	}
	loser := winner
	loser.ExpectedVersion = 0
	loser.Version = 2
	loser.TaskRequestReceipts = map[string]AIAgentTaskActionResponse{"task": {AssignmentID: "loser"}}
	if err := f.store.SaveAIAgentClientSnapshot(context.Background(), loser); !errors.Is(err, errAIAgentSnapshotConflict) {
		t.Fatalf("stale writer accepted: %v", err)
	}
	loaded, ok, err := f.store.LoadAIAgentClientSnapshot(context.Background())
	if err != nil || !ok || loaded.TaskRequestReceipts["task"].AssignmentID != "winner" {
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
