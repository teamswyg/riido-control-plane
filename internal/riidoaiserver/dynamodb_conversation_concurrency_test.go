package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

func TestConversationDynamoDBRejectsStaleOrdinaryWriter(t *testing.T) {
	f := newSnapshotDynamoDBFixture(t, fixedSnapshotTestNow(), nil, nil)
	defer f.close()
	before := conversationSnapshotForTest(101)
	if err := f.store.SaveAIAgentClientSnapshot(t.Context(), before); err != nil {
		t.Fatal(err)
	}
	stale := newSnapshotDynamoDBStore(t, fixedSnapshotTestNow(), f.server.URL, nil)
	defer stale.Close()
	if _, _, err := stale.LoadAIAgentClientSnapshot(t.Context()); err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(t.Context(), taskAgentPreserveMembershipContextKey{}, true)
	after := conversationSnapshotForTest(102)
	after.Version, after.ExpectedVersion = 2, 1
	if err := f.store.SaveAIAgentClientSnapshot(ctx, after); err != nil {
		t.Fatal(err)
	}
	before.Version, before.ExpectedVersion = 2, 1
	if err := stale.SaveAIAgentClientSnapshot(ctx, before); !errors.Is(err, errTaskAgentSnapshotChanged) {
		t.Fatalf("stale heartbeat overwrote newer history: %v", err)
	}
	got, found, err := stale.LoadAIAgentClientSnapshot(t.Context())
	if err != nil || !found || len(got.TaskThreadMessages["thread"]) != 102 {
		t.Fatalf("reload lost newer message: %t %v", found, err)
	}
	got.Version, got.ExpectedVersion = 3, 2
	if err := stale.SaveAIAgentClientSnapshot(ctx, got); err != nil {
		t.Fatalf("fresh retry: %v", err)
	}
}
