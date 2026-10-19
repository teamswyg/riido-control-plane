package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

func TestCredentialDeviceConnectPersistsAccountWithoutRotatingSecret(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatal(err)
	}
	owner := AuthorizationResult{PrincipalID: "owner", WorkspaceID: "owner-workspace"}
	guest := AuthorizationResult{PrincipalID: "guest", WorkspaceID: "guest-workspace"}
	credential, err := store.EnrollDeviceCredential(ctx, owner, owner.WorkspaceID, EnrollDeviceRequest{MachineID: "original-machine"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.ConnectAIAgentDeviceWithCredential(ctx, guest, credential.DeviceID, "wrong"); !errors.Is(err, ErrAuthorizationUnauthenticated) {
		t.Fatalf("invalid proof=%v", err)
	}
	if _, err = store.ConnectAIAgentDeviceWithCredential(ctx, guest, credential.DeviceID, credential.DeviceSecret); err != nil {
		t.Fatal(err)
	}
	reopened, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatal(err)
	}
	_, count := reopened.deviceConnectionSummaryLocked(credential.DeviceID)
	if count != 2 {
		t.Fatalf("connected principal count=%d", count)
	}
	if _, err = reopened.AuthorizeDeviceCredential(ctx, credential.DeviceID, credential.DeviceSecret, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionPoll}); err != nil {
		t.Fatal(err)
	}
	if reopened.deviceCredentials[credential.DeviceID].ownerPrincipalID != owner.PrincipalID {
		t.Fatal("device owner changed")
	}
}
