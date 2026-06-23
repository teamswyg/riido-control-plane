package riidoaiserver

import (
	"context"
	"testing"
)

func TestDeviceVisibleToConnectedWorkspaceMembers(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	const machine = "machine-uuid-shared"
	const ws = "workspace-shared"

	principal := AuthorizationResult{PrincipalID: "owner-user", WorkspaceID: ws}
	enroll := enrollTestDevice(t, ctx, store, principal, ws, machine, "Shared Mac")
	syncTestRuntime(t, ctx, store, principal, machine, enroll.DeviceID, machine+":claude")

	asMember, err := store.ListAIAgentDevices(ctx,
		AuthorizationResult{PrincipalID: "other-user", WorkspaceID: ws})
	if err != nil {
		t.Fatalf("list as member: %v", err)
	}
	if !containsDeviceID(asMember.Devices, enroll.DeviceID) {
		t.Fatalf("device %q not visible to connected workspace member", enroll.DeviceID)
	}

	asOther, err := store.ListAIAgentDevices(ctx,
		AuthorizationResult{PrincipalID: "other-user", WorkspaceID: "workspace-unconnected"})
	if err != nil {
		t.Fatalf("list unconnected workspace: %v", err)
	}
	if containsDeviceID(asOther.Devices, enroll.DeviceID) {
		t.Fatalf("device %q leaked into unconnected workspace", enroll.DeviceID)
	}
}
