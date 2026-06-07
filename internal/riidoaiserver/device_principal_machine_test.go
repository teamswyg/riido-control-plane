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

// A device is visible to every member of any workspace it is connected to
// (workspace-connection scoped), and NOT to unconnected workspaces (isolation).
func TestDeviceVisibleToConnectedWorkspaceMembers(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	const machine = "machine-uuid-shared"
	const ws = "workspace-shared"

	enroll, err := store.EnrollDeviceCredential(ctx,
		AuthorizationResult{PrincipalID: "owner-user", WorkspaceID: ws}, ws,
		EnrollDeviceRequest{MachineID: machine, DisplayName: "Shared Mac"})
	if err != nil {
		t.Fatalf("enroll: %v", err)
	}
	if _, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx,
		AuthorizationResult{PrincipalID: "owner-user", WorkspaceID: ws},
		DeviceRuntimeSnapshotSyncRequest{
			DaemonID: machine,
			DeviceID: enroll.DeviceID,
			Runtimes: []RuntimeSnapshotRecord{{
				RuntimeID:      machine + ":claude",
				Kind:           RuntimeKindClaudeCode,
				Availability:   RuntimeAvailabilityOnline,
				DetectionState: RuntimeDetectionStateDetected,
			}},
		}); err != nil {
		t.Fatalf("sync: %v", err)
	}

	// A DIFFERENT account that is a member of the connected workspace sees it.
	asMember, err := store.ListAIAgentDevices(ctx,
		AuthorizationResult{PrincipalID: "other-user", WorkspaceID: ws})
	if err != nil {
		t.Fatalf("list as member: %v", err)
	}
	if !containsDeviceID(asMember.Devices, enroll.DeviceID) {
		t.Fatalf("device %q not visible to member of connected workspace %q", enroll.DeviceID, ws)
	}

	// The same account in an unconnected workspace must NOT see it.
	asOther, err := store.ListAIAgentDevices(ctx,
		AuthorizationResult{PrincipalID: "other-user", WorkspaceID: "workspace-unconnected"})
	if err != nil {
		t.Fatalf("list unconnected ws: %v", err)
	}
	if containsDeviceID(asOther.Devices, enroll.DeviceID) {
		t.Fatalf("device %q leaked into unconnected workspace", enroll.DeviceID)
	}
}

// ConnectAIAgentDevice connects the machine's (shared) device to another
// workspace, so that workspace's members see the device with its runtimes —
// without re-enrolling or rotating the device secret.
func TestConnectAIAgentDeviceMakesDeviceVisibleInOtherWorkspace(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	const machine = "machine-uuid-multi"
	const wsA = "workspace-a"
	const wsB = "workspace-b"

	// jack-style owner enrolls + reports runtimes in workspace A.
	enroll, err := store.EnrollDeviceCredential(ctx,
		AuthorizationResult{PrincipalID: "owner-user", WorkspaceID: wsA}, wsA,
		EnrollDeviceRequest{MachineID: machine, DisplayName: "Multi Mac"})
	if err != nil {
		t.Fatalf("enroll: %v", err)
	}
	const runtimeID = machine + ":claude"
	if _, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx,
		AuthorizationResult{PrincipalID: "owner-user", WorkspaceID: wsA},
		DeviceRuntimeSnapshotSyncRequest{
			DaemonID: machine,
			DeviceID: enroll.DeviceID,
			Runtimes: []RuntimeSnapshotRecord{{
				RuntimeID:      runtimeID,
				Kind:           RuntimeKindClaudeCode,
				Availability:   RuntimeAvailabilityOnline,
				DetectionState: RuntimeDetectionStateDetected,
			}},
		}); err != nil {
		t.Fatalf("sync: %v", err)
	}

	// A different account, in workspace B, is not a member of A and does not see
	// the device yet.
	before, err := store.ListAIAgentDevices(ctx, AuthorizationResult{PrincipalID: "other-user", WorkspaceID: wsB})
	if err != nil {
		t.Fatalf("list before connect: %v", err)
	}
	if containsDeviceID(before.Devices, enroll.DeviceID) {
		t.Fatalf("device unexpectedly visible in %s before connect", wsB)
	}

	// Connecting the machine to workspace B (by machine_id) makes it visible to B.
	if _, err := store.ConnectAIAgentDevice(ctx,
		AuthorizationResult{PrincipalID: "other-user", WorkspaceID: wsB}, machine); err != nil {
		t.Fatalf("connect: %v", err)
	}
	after, err := store.ListAIAgentDevices(ctx, AuthorizationResult{PrincipalID: "other-user", WorkspaceID: wsB})
	if err != nil {
		t.Fatalf("list after connect: %v", err)
	}
	var mine *DeviceRecord
	for i := range after.Devices {
		if after.Devices[i].DeviceID == enroll.DeviceID {
			mine = &after.Devices[i]
			break
		}
	}
	if mine == nil {
		t.Fatalf("device %q not visible in %s after connect", enroll.DeviceID, wsB)
	}
	if len(mine.Runtimes) != 1 || mine.Runtimes[0].RuntimeID != runtimeID {
		t.Fatalf("connected device missing owner-reported runtimes: %+v", mine.Runtimes)
	}
}

// The daemon learns which agents to poll from /v1/daemon/agent-bindings. Because
// one physical machine serves many workspaces under a single device credential,
// the bindings must include agents created in ANY workspace the device is
// connected to — not only the credential's enroll workspace. Otherwise an agent
// assigned from another connected workspace is never polled and its assignment is
// stuck "queued" forever.
func TestDaemonAgentBindingsIncludeAgentsFromOtherConnectedWorkspaces(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	const machine = "machine-binding-multi"
	const wsA = "workspace-a"
	const wsB = "workspace-b"
	const owner = "owner-user"

	enroll, err := store.EnrollDeviceCredential(ctx,
		AuthorizationResult{PrincipalID: owner, WorkspaceID: wsA}, wsA,
		EnrollDeviceRequest{MachineID: machine, DisplayName: "Multi Mac"})
	if err != nil {
		t.Fatalf("enroll: %v", err)
	}
	runtimeID := machine + ":claude"
	if _, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx,
		AuthorizationResult{PrincipalID: owner, WorkspaceID: wsA},
		DeviceRuntimeSnapshotSyncRequest{
			DaemonID: machine,
			DeviceID: enroll.DeviceID,
			Runtimes: []RuntimeSnapshotRecord{{
				RuntimeID:      runtimeID,
				Kind:           RuntimeKindClaudeCode,
				Availability:   RuntimeAvailabilityOnline,
				DetectionState: RuntimeDetectionStateDetected,
			}},
		}); err != nil {
		t.Fatalf("sync: %v", err)
	}

	// Connect the machine to workspace B and create an agent there bound to the
	// device's runtime — an agent whose workspace differs from the credential's.
	if _, err := store.ConnectAIAgentDevice(ctx,
		AuthorizationResult{PrincipalID: owner, WorkspaceID: wsB}, machine); err != nil {
		t.Fatalf("connect wsB: %v", err)
	}
	created, err := store.createAIAgent(ctx,
		AuthorizationResult{PrincipalID: owner, WorkspaceID: wsB},
		CreateAgentConfigurationRequest{Name: "B Agent", RuntimeID: runtimeID, Visibility: AgentVisibilityPrivate}, "")
	if err != nil {
		t.Fatalf("create agent in wsB: %v", err)
	}

	// The daemon authenticates with the device credential (enroll workspace wsA, no
	// explicit workspace). Its bindings must still include the wsB agent.
	bindings, err := store.ListAIAgentDaemonAgentBindings(ctx,
		AuthorizationResult{PrincipalID: owner}, enroll.DeviceID)
	if err != nil {
		t.Fatalf("list bindings: %v", err)
	}
	var found *AgentRuntimeBinding
	for i := range bindings.Bindings {
		if bindings.Bindings[i].AgentID == created.Agent.AgentID {
			found = &bindings.Bindings[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("agent in connected workspace %s missing from daemon bindings: %+v", wsB, bindings.Bindings)
	}
	if found.RuntimeID != runtimeID || found.DeviceID != enroll.DeviceID {
		t.Fatalf("binding has wrong runtime/device: %+v", *found)
	}
}

func containsDeviceID(devices []DeviceRecord, deviceID string) bool {
	for _, d := range devices {
		if d.DeviceID == deviceID {
			return true
		}
	}
	return false
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
