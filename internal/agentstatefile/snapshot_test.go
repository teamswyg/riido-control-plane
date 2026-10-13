package agentstatefile

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestFileAgentSnapshotReopenAndVersionFence(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "agents.json")
	disk := NewFileAIAgentClientSnapshot(path)
	_, err := riidoaiserver.OpenPersistentAIAgentClientStore(ctx, nil, disk)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, ok, err := disk.LoadAIAgentClientSnapshot(ctx)
	if err != nil || !ok {
		t.Fatalf("snapshot missing: %v", err)
	}
	if snapshot.Version <= 0 {
		t.Fatal("persisted version mismatch")
	}
	if _, err := riidoaiserver.OpenPersistentAIAgentClientStore(ctx, nil, NewFileAIAgentClientSnapshot(path)); err != nil {
		t.Fatal(err)
	}
	snapshot.ExpectedVersion = snapshot.Version - 1
	if err := disk.SaveAIAgentClientSnapshot(ctx, snapshot); err == nil {
		t.Fatal("stale write accepted")
	}
}
