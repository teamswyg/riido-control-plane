package riidoaiserver

import (
	"context"
	"testing"
)

// One machine must resolve to exactly one device row regardless of how many
// workspaces it enrolls in: enrolling the same machine adds one row, not one per
// (workspace) enrollment.
func TestEnrollDeviceCredentialIsIdempotentPerMachine(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	baseline := len(store.devices)
	const machine = "machine-uuid-abc"

	first, err := store.EnrollDeviceCredential(ctx,
		AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-a"}, "workspace-a",
		EnrollDeviceRequest{MachineID: machine, DisplayName: "MacBook Pro SK"})
	if err != nil {
		t.Fatalf("enroll workspace-a: %v", err)
	}
	second, err := store.EnrollDeviceCredential(ctx,
		AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-b"}, "workspace-b",
		EnrollDeviceRequest{MachineID: machine, DisplayName: "MacBook Pro SK"})
	if err != nil {
		t.Fatalf("enroll workspace-b: %v", err)
	}
	if first.DeviceID != second.DeviceID {
		t.Fatalf("same machine across workspaces must yield same DeviceID: %q vs %q", first.DeviceID, second.DeviceID)
	}

	other, err := store.EnrollDeviceCredential(ctx,
		AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-a"}, "workspace-a",
		EnrollDeviceRequest{MachineID: "machine-uuid-xyz", DisplayName: "Other Mac"})
	if err != nil {
		t.Fatalf("enroll other machine: %v", err)
	}
	if other.DeviceID == first.DeviceID {
		t.Fatal("different machines must yield different DeviceIDs")
	}

	if got := len(store.devices) - baseline; got != 2 {
		t.Fatalf("added device rows = %d, want 2 (one per machine, not per enrollment)", got)
	}
}

// Stale legacy runtimes (agentd-local:*) minted before per-machine UUIDs must be
// dropped on restore; real per-machine runtimes are kept.
func TestPruneLegacyRuntimeRecordsDropsAgentdLocal(t *testing.T) {
	devices := []DeviceRecord{{
		DeviceID: "dev_x",
		Runtimes: []RuntimeRecord{
			{RuntimeID: "agentd-local:claude"},
			{RuntimeID: "e98eefcd-66b4-49e2-a1bf-1cab74749e2d:claude"},
			{RuntimeID: "agentd-local:codex"},
		},
	}}
	out := pruneLegacyRuntimeRecords(devices)
	if len(out[0].Runtimes) != 1 ||
		out[0].Runtimes[0].RuntimeID != "e98eefcd-66b4-49e2-a1bf-1cab74749e2d:claude" {
		t.Fatalf("legacy runtimes not pruned: %+v", out[0].Runtimes)
	}
}

// A re-enroll (workspace switch / credential refresh) must not create a new row
// nor wipe the runtimes the daemon already reported for that device.
func TestEnrollDeviceCredentialReEnrollPreservesRuntimes(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	baseline := len(store.devices)
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-a"}
	const machine = "machine-uuid-abc"

	enroll, err := store.EnrollDeviceCredential(ctx, principal, "workspace-a",
		EnrollDeviceRequest{MachineID: machine, DisplayName: "MacBook Pro SK"})
	if err != nil {
		t.Fatalf("enroll: %v", err)
	}
	const runtimeID = "machine-uuid-abc:claude"
	if _, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, DeviceRuntimeSnapshotSyncRequest{
		DaemonID: machine,
		DeviceID: enroll.DeviceID,
		Runtimes: []RuntimeSnapshotRecord{{
			RuntimeID:      runtimeID,
			Kind:           RuntimeKindClaudeCode,
			Availability:   RuntimeAvailabilityOnline,
			DetectionState: RuntimeDetectionStateDetected,
		}},
	}); err != nil {
		t.Fatalf("sync runtime snapshot: %v", err)
	}

	// Re-enroll the same machine from a different workspace.
	if _, err := store.EnrollDeviceCredential(ctx,
		AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-b"}, "workspace-b",
		EnrollDeviceRequest{MachineID: machine, DisplayName: "MacBook Pro SK"}); err != nil {
		t.Fatalf("re-enroll: %v", err)
	}

	if got := len(store.devices) - baseline; got != 1 {
		t.Fatalf("added device rows = %d, want 1 (re-enroll must not add a row)", got)
	}

	devices, err := store.ListAIAgentDevices(ctx, principal)
	if err != nil {
		t.Fatalf("list devices: %v", err)
	}
	var mine *DeviceRecord
	for i := range devices.Devices {
		if devices.Devices[i].DeviceID == enroll.DeviceID {
			mine = &devices.Devices[i]
			break
		}
	}
	if mine == nil {
		t.Fatalf("enrolled device %q not visible to owner", enroll.DeviceID)
	}
	if len(mine.Runtimes) != 1 || mine.Runtimes[0].RuntimeID != runtimeID {
		t.Fatalf("re-enroll wiped detected runtimes: %+v", mine.Runtimes)
	}
}
