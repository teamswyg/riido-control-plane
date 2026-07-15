package riidoaiserver

import (
	"context"
	"testing"
)

func TestDeviceConnectionGrantsSurviveSnapshotAndKeepOriginalCredential(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	writer, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("open writer: %v", err)
	}
	owner := AuthorizationResult{PrincipalID: "owner-account", WorkspaceID: "workspace-owner"}
	guest := AuthorizationResult{PrincipalID: "guest-account", WorkspaceID: "workspace-guest"}
	enrollment, err := writer.EnrollDeviceCredential(ctx, owner, owner.WorkspaceID, EnrollDeviceRequest{
		MachineID:   "machine-cross-account-persisted",
		DisplayName: "Persistent Shared Mac",
	})
	if err != nil {
		t.Fatalf("enroll: %v", err)
	}
	if _, err := writer.ConnectAIAgentDevice(ctx, guest, "machine-cross-account-persisted"); err != nil {
		t.Fatalf("connect guest: %v", err)
	}

	reopened, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	reopened.mu.Lock()
	revision, principalCount := reopened.deviceConnectionSummaryLocked(enrollment.DeviceID)
	reopened.mu.Unlock()
	if principalCount != 2 || revision == "" {
		t.Fatalf("restored connection summary count=%d revision=%q", principalCount, revision)
	}
	if _, err := reopened.AuthorizeDeviceCredential(ctx, enrollment.DeviceID, enrollment.DeviceSecret, AuthorizationRequest{
		Resource: AuthorizationResourceAgent,
		Action:   AuthorizationActionPoll,
	}); err != nil {
		t.Fatalf("original credential did not survive account connection: %v", err)
	}
	devices, err := reopened.ListAIAgentDevices(ctx, guest)
	if err != nil {
		t.Fatalf("list guest devices: %v", err)
	}
	device := requireDevice(t, devices.Devices, enrollment.DeviceID)
	if device.OwnerPrincipalID != owner.PrincipalID {
		t.Fatalf("restored owner = %q, want %q", device.OwnerPrincipalID, owner.PrincipalID)
	}
}
