package riidoaiserver

import (
	"context"
	"path/filepath"
	"testing"
)

func TestFileAgentSnapshotReopenAndVersionFence(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "agents.json")
	disk := NewFileAIAgentClientSnapshot(path)
	store, err := OpenPersistentAIAgentClientStore(ctx, nil, disk)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, ok, err := disk.LoadAIAgentClientSnapshot(ctx)
	if err != nil || !ok {
		t.Fatalf("snapshot missing: %v", err)
	}
	if snapshot.Version != store.version {
		t.Fatal("persisted version mismatch")
	}
	if _, err := OpenPersistentAIAgentClientStore(ctx, nil, NewFileAIAgentClientSnapshot(path)); err != nil {
		t.Fatal(err)
	}
	snapshot.ExpectedVersion = snapshot.Version - 1
	if err := disk.SaveAIAgentClientSnapshot(ctx, snapshot); err == nil {
		t.Fatal("stale write accepted")
	}
}
