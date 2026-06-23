package riidoaiserver

import (
	"context"
	"testing"
)

func TestEnrollDeviceCredentialIsIdempotentPerMachine(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	baseline := len(store.devices)
	const machine = "machine-uuid-abc"

	first := enrollTestDevice(t, ctx, store,
		AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-a"},
		"workspace-a", machine, "MacBook Pro SK")
	second := enrollTestDevice(t, ctx, store,
		AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-b"},
		"workspace-b", machine, "MacBook Pro SK")
	if first.DeviceID != second.DeviceID {
		t.Fatalf("same machine yielded different DeviceIDs: %q vs %q", first.DeviceID, second.DeviceID)
	}

	other := enrollTestDevice(t, ctx, store,
		AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-a"},
		"workspace-a", "machine-uuid-xyz", "Other Mac")
	if other.DeviceID == first.DeviceID {
		t.Fatal("different machines must yield different DeviceIDs")
	}
	if got := len(store.devices) - baseline; got != 2 {
		t.Fatalf("added device rows = %d, want 2", got)
	}
}

func TestEnrollDeviceCredentialReEnrollPreservesRuntimes(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	baseline := len(store.devices)
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-a"}
	const machine = "machine-uuid-abc"
	const runtimeID = "machine-uuid-abc:claude"

	enroll := enrollTestDevice(t, ctx, store, principal, "workspace-a", machine, "MacBook Pro SK")
	syncTestRuntime(t, ctx, store, principal, machine, enroll.DeviceID, runtimeID)
	if _, err := store.EnrollDeviceCredential(ctx,
		AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-b"}, "workspace-b",
		EnrollDeviceRequest{MachineID: machine, DisplayName: "MacBook Pro SK"}); err != nil {
		t.Fatalf("re-enroll: %v", err)
	}

	if got := len(store.devices) - baseline; got != 1 {
		t.Fatalf("added device rows = %d, want 1", got)
	}
	devices, err := store.ListAIAgentDevices(ctx, principal)
	if err != nil {
		t.Fatalf("list devices: %v", err)
	}
	mine := requireDevice(t, devices.Devices, enroll.DeviceID)
	if len(mine.Runtimes) != 1 || mine.Runtimes[0].RuntimeID != runtimeID {
		t.Fatalf("re-enroll wiped detected runtimes: %+v", mine.Runtimes)
	}
}
