package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

type failingPresenceCommitStore struct {
	grantsReadingSnapshotStore
	fail bool
}

func (s *failingPresenceCommitStore) SaveAIAgentClientSnapshot(ctx context.Context, snapshot AIAgentClientSnapshot) error {
	if s.fail {
		return errors.New("injected commit failure")
	}
	return s.memoryAIAgentClientSnapshotStore.SaveAIAgentClientSnapshot(ctx, snapshot)
}

func TestDesktopPresenceFailedConnectionDoesNotPublishOnLaterSave(t *testing.T) {
	for _, enroll := range []bool{false, true} {
		t.Run(map[bool]string{false: "connect", true: "enroll"}[enroll], func(t *testing.T) {
			ctx := context.Background()
			principal := profileColorTestPrincipal()
			snapshots := &failingPresenceCommitStore{grantsReadingSnapshotStore{&memoryAIAgentClientSnapshotStore{}}, false}
			store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := store.ConnectAIAgentDevice(ctx, principal, "previously-committed"); err != nil {
				t.Fatal(err)
			}
			baseline := store.desktopPresenceGrantCheckpoint()
			snapshots.fail = true
			if enroll {
				_, err = store.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, EnrollDeviceRequest{MachineID: "failed-new-device"})
			} else {
				_, err = store.ConnectAIAgentDevice(ctx, principal, "failed-new-device")
			}
			if err == nil {
				t.Fatal("expected commit failure")
			}
			assertSameJSON(t, baseline, store.desktopPresenceGrantCheckpoint())
			snapshots.fail = false
			if _, err = store.UpdateAIAgentConfiguration(ctx, principal, "agent-owned-claude", UpdateAgentConfigurationRequest{Name: "Later successful update"}); err != nil {
				t.Fatal(err)
			}
			// Exercise the inline/file JSON representation as well as split parts.
			serialized, marshalErr := json.Marshal(snapshots.snapshot)
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			if unmarshalErr := json.Unmarshal(serialized, &snapshots.snapshot); unmarshalErr != nil {
				t.Fatal(unmarshalErr)
			}
			fresh, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
			if err != nil {
				t.Fatal(err)
			}
			assertSameJSON(t, baseline, fresh.desktopPresenceGrantCheckpoint())
			if err = fresh.saveSnapshot(ctx); err != nil {
				t.Fatal(err)
			}
			assertSameJSON(t, baseline, snapshots.snapshot.DeviceConnectionGrants)
		})
	}
}
