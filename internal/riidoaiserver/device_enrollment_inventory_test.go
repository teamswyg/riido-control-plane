package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestDynamoEnrollmentPreservesCommittedHeartbeatInventory(t *testing.T) {
	for _, sameDevice := range []bool{true, false} {
		t.Run(map[bool]string{true: "same-device", false: "other-device"}[sameDevice], func(t *testing.T) {
			ctx, principal := context.Background(), profileColorTestPrincipal()
			fixture := newMembershipSharedDynamoFixture(t)
			first, native := openEnrollmentInventoryStore(t, fixture.server.URL)
			req := EnrollDeviceRequest{MachineID: "runtime-inventory"}
			enrollment, err := first.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, req)
			if err != nil {
				t.Fatal(err)
			}
			stale, _ := openEnrollmentInventoryStore(t, fixture.server.URL)
			// Keep this replica deliberately stale to exercise the storage merge.
			stale.snapshotReloadInterval = time.Hour
			before, _ := stale.deviceByIDLocked(enrollment.DeviceID)
			committed, found, err := native.LoadAIAgentClientSnapshot(ctx)
			if err != nil || !found {
				t.Fatal("committed snapshot missing")
			}
			heartbeat := time.Now().UTC()
			for i := range committed.Devices {
				if committed.Devices[i].DeviceID == enrollment.DeviceID {
					committed.Devices[i].DaemonLastSeenAt = heartbeat
					committed.Devices[i].Runtimes = []RuntimeRecord{{RuntimeID: "fresh-runtime-from-other-replica"}}
				}
			}
			preserve := context.WithValue(ctx, taskAgentPreserveMembershipContextKey{}, true)
			if err := native.SaveAIAgentClientSnapshot(preserve, committed); err != nil {
				t.Fatal(err)
			}
			if !sameDevice {
				req.MachineID = "different-enrolled-machine"
			}
			rotated, err := stale.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, req)
			if err != nil {
				t.Fatal(err)
			}
			if !rotated.IssuedAt.After(heartbeat) {
				t.Fatal("regression requires enrollment after remote heartbeat")
			}
			assertEnrollmentHeartbeatInventory(t, native, enrollment.DeviceID, heartbeat)
			local, _ := stale.deviceByIDLocked(enrollment.DeviceID)
			if !local.DaemonLastSeenAt.Equal(before.DaemonLastSeenAt) {
				t.Fatal("enrollment advanced stale local heartbeat")
			}
			if _, err := stale.UpdateAIAgentConfiguration(ctx, principal, "agent-owned-claude", UpdateAgentConfigurationRequest{Name: "ordinary save after enrollment"}); err != nil {
				t.Fatal(err)
			}
			assertEnrollmentHeartbeatInventory(t, native, enrollment.DeviceID, heartbeat)
		})
	}
}
