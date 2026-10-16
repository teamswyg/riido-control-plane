package riidoaiserver

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDesktopPresenceEmptyGrantCommitRemainsEmptyAfterRestart(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	snapshots := &failingPresenceCommitStore{grantsReadingSnapshotStore{&memoryAIAgentClientSnapshotStore{}}, false}
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatal(err)
	}
	snapshots.fail = true
	if _, err = store.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, EnrollDeviceRequest{MachineID: "failed-only-device"}); err == nil {
		t.Fatal("expected enrollment save failure")
	}
	snapshots.fail = false
	if _, err = store.UpdateAIAgentConfiguration(ctx, principal, "agent-owned-claude", UpdateAgentConfigurationRequest{Name: "Unrelated commit"}); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(snapshots.snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var persisted AIAgentClientSnapshot
	if err = json.Unmarshal(raw, &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted.DeviceConnectionGrants == nil {
		t.Fatal("empty modern grants lost their explicit representation")
	}
	snapshots.snapshot = persisted
	fresh, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatal(err)
	}
	if got := fresh.desktopPresenceGrantCheckpoint(); len(got) != 0 {
		t.Fatal("failed enrollment resurrected as a legacy connection")
	}
	if seen, err := fresh.DesktopPresenceLastSeenAt(ctx, principal); err != nil || !seen.IsZero() {
		t.Fatalf("failed enrollment appeared after restart: %v", err)
	}
}
