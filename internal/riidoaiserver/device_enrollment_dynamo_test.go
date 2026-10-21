package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestDynamoEnrollmentPreservesRotationAcrossStaleOrdinaryWrites(t *testing.T) {
	fixture := newMembershipSharedDynamoFixture(t)
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	open := func() *PersistentAIAgentClientStore {
		snapshots := newSnapshotDynamoDBStore(t, time.Now(), fixture.server.URL, nil)
		t.Cleanup(func() { _ = snapshots.Close() })
		store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
		if err != nil {
			t.Fatal(err)
		}
		return store
	}
	first := open()
	req := EnrollDeviceRequest{MachineID: "native-rotated-device", DisplayName: "Original"}
	old, err := first.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, req)
	if err != nil {
		t.Fatal(err)
	}
	stale := open()
	req.DisplayName, req.AppVersion = "Rotated", "9.8.7"
	rotated, err := first.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, req)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stale.UpdateAIAgentConfiguration(ctx, principal, "agent-owned-claude", UpdateAgentConfigurationRequest{Name: "stale ordinary writer"}); err != nil {
		t.Fatal(err)
	}
	fresh := open()
	assertEnrollmentCredential(t, fresh, rotated, true)
	assertEnrollmentCredential(t, fresh, old, false)
	device, found := fresh.deviceByIDLocked(rotated.DeviceID)
	if !found || device.DisplayName != "Rotated" || device.DesktopAppVersion != "9.8.7" {
		t.Fatal("stale save rewrote enrollment-owned device fields")
	}
	// A stale other replica may enroll a different device without erasing this one.
	other, err := stale.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, EnrollDeviceRequest{MachineID: "native-other-device"})
	if err != nil {
		t.Fatal(err)
	}
	final := open()
	assertEnrollmentCredential(t, final, rotated, true)
	assertEnrollmentCredential(t, final, other, true)
}

func TestDynamoConcurrentDisjointEnrollmentsKeepBothCredentials(t *testing.T) {
	ctx := context.Background()
	fixture := newMembershipSharedDynamoFixture(t)
	principal := profileColorTestPrincipal()
	var stores []*PersistentAIAgentClientStore
	for range 2 {
		native := newSnapshotDynamoDBStore(t, time.Now(), fixture.server.URL, nil)
		t.Cleanup(func() { _ = native.Close() })
		store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), native)
		if err != nil {
			t.Fatal(err)
		}
		stores = append(stores, store)
	}
	type result struct {
		credential EnrollDeviceResponse
		err        error
	}
	results, start := make(chan result, 2), make(chan struct{})
	for i, store := range stores {
		go func(i int, store *PersistentAIAgentClientStore) {
			<-start
			credential, err := store.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, EnrollDeviceRequest{MachineID: []string{"concurrent-device-one", "concurrent-device-two"}[i]})
			results <- result{credential, err}
		}(i, store)
	}
	close(start)
	first, second := <-results, <-results
	if first.err != nil || second.err != nil {
		t.Fatalf("disjoint enrollment failed: %v %v", first.err, second.err)
	}
	native := newSnapshotDynamoDBStore(t, time.Now(), fixture.server.URL, nil)
	defer native.Close()
	fresh, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), native)
	if err != nil {
		t.Fatal(err)
	}
	assertEnrollmentCredential(t, fresh, first.credential, true)
	assertEnrollmentCredential(t, fresh, second.credential, true)
}
