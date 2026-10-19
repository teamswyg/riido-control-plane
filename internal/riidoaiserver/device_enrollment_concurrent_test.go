package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

type enrollmentBlockingSnapshotStore struct {
	AIAgentClientSnapshotStore
	entered, release chan struct{}
}

func (s *enrollmentBlockingSnapshotStore) SaveAIAgentClientSnapshot(ctx context.Context, snapshot AIAgentClientSnapshot) error {
	if _, enrolling := ctx.Value(deviceEnrollmentMutationContextKey{}).(deviceEnrollmentMutation); enrolling && s.entered != nil {
		close(s.entered)
		<-s.release
	}
	return s.AIAgentClientSnapshotStore.SaveAIAgentClientSnapshot(ctx, snapshot)
}

func TestEnrollmentPublicationPreservesHeartbeatAndOtherPrincipalGrant(t *testing.T) {
	ctx := context.Background()
	fixture := newMembershipSharedDynamoFixture(t)
	open := func() (*PersistentAIAgentClientStore, *enrollmentBlockingSnapshotStore) {
		native := newSnapshotDynamoDBStore(t, time.Now(), fixture.server.URL, nil)
		t.Cleanup(func() { _ = native.Close() })
		blocked := &enrollmentBlockingSnapshotStore{AIAgentClientSnapshotStore: native}
		store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), blocked)
		if err != nil {
			t.Fatal(err)
		}
		return store, blocked
	}
	store, blocking := open()
	principal := profileColorTestPrincipal()
	req := EnrollDeviceRequest{MachineID: "heartbeat-preservation"}
	first, err := store.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, req)
	if err != nil {
		t.Fatal(err)
	}
	peer, _ := open()
	blocking.entered, blocking.release = make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := store.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, req)
		done <- err
	}()
	<-blocking.entered
	heartbeat := time.Now().UTC()
	store.mu.Lock()
	for i := range store.devices {
		if store.devices[i].DeviceID == first.DeviceID {
			store.devices[i].DaemonLastSeenAt = heartbeat
			store.devices[i].Runtimes = []RuntimeRecord{{RuntimeID: "new-heartbeat-runtime"}}
		}
	}
	store.mu.Unlock()
	other := AuthorizationResult{PrincipalID: "other-principal", WorkspaceID: "other-workspace"}
	if _, err := peer.ConnectAIAgentDevice(ctx, other, req.MachineID); err != nil {
		t.Fatal(err)
	}
	close(blocking.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	store.mu.Lock()
	device, _ := store.deviceByIDLocked(first.DeviceID)
	store.mu.Unlock()
	if !device.DaemonLastSeenAt.Equal(heartbeat) || len(device.Runtimes) != 1 || device.Runtimes[0].RuntimeID != "new-heartbeat-runtime" {
		t.Fatal("publication restored stale heartbeat state")
	}
	fresh, _ := open()
	grants := fresh.desktopPresenceGrantCheckpoint()
	seen := false
	for _, grant := range grants {
		if grant.PrincipalID == other.PrincipalID && grant.WorkspaceID == other.WorkspaceID {
			seen = true
		}
	}
	if !seen {
		t.Fatal("enrollment erased a concurrent other-principal connection")
	}
}
