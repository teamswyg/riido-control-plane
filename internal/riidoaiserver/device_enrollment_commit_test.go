package riidoaiserver

import (
	"context"
	"reflect"
	"testing"
)

func TestFailedEnrollmentPreservesCommittedDeviceState(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(map[bool]string{false: "new", true: "existing"}[existing], func(t *testing.T) {
			ctx := context.Background()
			principal := profileColorTestPrincipal()
			snapshots := &failingPresenceCommitStore{grantsReadingSnapshotStore{&memoryAIAgentClientSnapshotStore{}}, false}
			store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
			if err != nil {
				t.Fatal(err)
			}
			req := EnrollDeviceRequest{MachineID: "transactional-enrollment"}
			var first EnrollDeviceResponse
			if existing {
				first, err = store.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, req)
				if err != nil {
					t.Fatal(err)
				}
			}
			before, err := store.snapshot(store.snapshotNow())
			if err != nil {
				t.Fatal(err)
			}
			snapshots.fail = true
			if _, err = store.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, req); err == nil {
				t.Fatal("expected write failure")
			}
			after, err := store.snapshot(store.snapshotNow())
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before.DeviceCredentials, after.DeviceCredentials) {
				t.Fatal("failed enrollment published credential state")
			}
			assertSameJSON(t, before.Devices, after.Devices)
			assertSameJSON(t, before.DeviceConnectionGrants, after.DeviceConnectionGrants)
			snapshots.fail = false
			if _, err = store.UpdateAIAgentConfiguration(ctx, principal, "agent-owned-claude", UpdateAgentConfigurationRequest{Name: "unrelated save"}); err != nil {
				t.Fatal(err)
			}
			fresh, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
			if err != nil {
				t.Fatal(err)
			}
			if existing {
				assertEnrollmentCredential(t, fresh, first, true)
			}
			if !reflect.DeepEqual(before.DeviceCredentials, snapshots.snapshot.DeviceCredentials) {
				t.Fatal("later save persisted failed credential state")
			}
		})
	}
}

func assertEnrollmentCredential(t *testing.T, store *PersistentAIAgentClientStore, enrollment EnrollDeviceResponse, accepted bool) {
	t.Helper()
	_, err := store.AuthorizeDeviceCredential(context.Background(), enrollment.DeviceID, enrollment.DeviceSecret, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionProviderStatusWrite, DeviceID: enrollment.DeviceID})
	if (err == nil) != accepted {
		t.Fatalf("credential acceptance=%v want=%v", err == nil, accepted)
	}
}
