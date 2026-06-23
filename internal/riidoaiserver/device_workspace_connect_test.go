package riidoaiserver

import (
	"context"
	"testing"
)

func TestConnectAIAgentDeviceMakesDeviceVisibleInOtherWorkspace(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	const machine = "machine-uuid-multi"
	const wsA = "workspace-a"
	const wsB = "workspace-b"
	const runtimeID = machine + ":claude"

	owner := AuthorizationResult{PrincipalID: "owner-user", WorkspaceID: wsA}
	enroll := enrollTestDevice(t, ctx, store, owner, wsA, machine, "Multi Mac")
	syncTestRuntime(t, ctx, store, owner, machine, enroll.DeviceID, runtimeID)

	before, err := store.ListAIAgentDevices(ctx,
		AuthorizationResult{PrincipalID: "other-user", WorkspaceID: wsB})
	if err != nil {
		t.Fatalf("list before connect: %v", err)
	}
	if containsDeviceID(before.Devices, enroll.DeviceID) {
		t.Fatalf("device unexpectedly visible in %s before connect", wsB)
	}
	if _, err := store.ConnectAIAgentDevice(ctx,
		AuthorizationResult{PrincipalID: "other-user", WorkspaceID: wsB}, machine); err != nil {
		t.Fatalf("connect: %v", err)
	}

	after, err := store.ListAIAgentDevices(ctx,
		AuthorizationResult{PrincipalID: "other-user", WorkspaceID: wsB})
	if err != nil {
		t.Fatalf("list after connect: %v", err)
	}
	mine := requireDevice(t, after.Devices, enroll.DeviceID)
	if len(mine.Runtimes) != 1 || mine.Runtimes[0].RuntimeID != runtimeID {
		t.Fatalf("connected device missing owner-reported runtimes: %+v", mine.Runtimes)
	}
}
