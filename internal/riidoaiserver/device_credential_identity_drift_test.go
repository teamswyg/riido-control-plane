package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

func TestEnrollmentKeepsVerifiedDeviceAcrossMachineIDDrift(t *testing.T) {
	s := NewDevelopmentAIAgentClientStore()
	ctx := context.Background()
	owner := AuthorizationResult{PrincipalID: "owner", WorkspaceID: "workspace"}
	first, err := s.EnrollDeviceCredential(ctx, owner, owner.WorkspaceID, EnrollDeviceRequest{MachineID: "original", DisplayName: "Mac"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.EnrollDeviceCredential(ctx, owner, owner.WorkspaceID, EnrollDeviceRequest{MachineID: "changed-local-file", DeviceID: first.DeviceID, DisplayName: "Mac"})
	if err != nil {
		t.Fatal(err)
	}
	if first.DeviceID != second.DeviceID || s.deviceCredentials[first.DeviceID].machineID != "original" {
		t.Fatal("created duplicate or replaced durable machine identity")
	}
	intruder := AuthorizationResult{PrincipalID: "other", WorkspaceID: "workspace"}
	if _, err := s.EnrollDeviceCredential(ctx, intruder, intruder.WorkspaceID, EnrollDeviceRequest{MachineID: "another", DeviceID: first.DeviceID}); !errors.Is(err, ErrAuthorizationForbidden) {
		t.Fatalf("cross-owner enrollment: %v", err)
	}
	other, err := s.EnrollDeviceCredential(ctx, owner, owner.WorkspaceID, EnrollDeviceRequest{MachineID: "other-machine", DisplayName: "Mac"})
	if err != nil {
		t.Fatal(err)
	}
	if first.DeviceID == other.DeviceID {
		t.Fatal("merged different machines by display name")
	}
}
func TestRuntimeModelSnapshotDoesNotInventDefault(t *testing.T) {
	models := normalizeRuntimeModels(RuntimeKindCursor, []RuntimeModelRecord{{ModelID: "a"}, {ModelID: "cursor-auto"}})
	for _, model := range models {
		if model.IsDefault {
			t.Fatal("invented default")
		}
	}
}
