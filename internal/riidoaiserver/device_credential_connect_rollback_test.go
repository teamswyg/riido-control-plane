package riidoaiserver

import (
	"context"
	"testing"
)

func TestCredentialDeviceConnectFailedSaveRestoresWorkspace(t *testing.T) {
	ctx := context.Background()
	snapshots := &failingPresenceCommitStore{grantsReadingSnapshotStore{&memoryAIAgentClientSnapshotStore{}}, false}
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatal(err)
	}
	owner := AuthorizationResult{PrincipalID: "owner", WorkspaceID: "owner-workspace"}
	guest := AuthorizationResult{PrincipalID: "guest", WorkspaceID: "rejected-workspace"}
	credential, err := store.EnrollDeviceCredential(ctx, owner, owner.WorkspaceID, EnrollDeviceRequest{MachineID: "rollback-device"})
	if err != nil {
		t.Fatal(err)
	}
	before := store.desktopPresenceGrantCheckpoint()
	snapshots.fail = true
	if _, err = store.ConnectAIAgentDeviceWithCredential(ctx, guest, credential.DeviceID, credential.DeviceSecret); err == nil {
		t.Fatal("expected failed save")
	}
	assertSameJSON(t, before, store.desktopPresenceGrantCheckpoint())
	assertWorkspaceAbsent := func(s *PersistentAIAgentClientStore) {
		t.Helper()
		s.mu.Lock()
		defer s.mu.Unlock()
		for _, device := range s.devices {
			if device.DeviceID != credential.DeviceID {
				continue
			}
			for _, workspace := range device.ConnectedWorkspaceIDs {
				if workspace == guest.WorkspaceID {
					t.Fatal("rejected workspace connection survived failed save")
				}
			}
		}
	}
	assertWorkspaceAbsent(store)
	snapshots.fail = false
	if err = store.saveSnapshot(ctx); err != nil {
		t.Fatal(err)
	}
	reopened, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatal(err)
	}
	assertWorkspaceAbsent(reopened)
	assertSameJSON(t, before, reopened.desktopPresenceGrantCheckpoint())
}
