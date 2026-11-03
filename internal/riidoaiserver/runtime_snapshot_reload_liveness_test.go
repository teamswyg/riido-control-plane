package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

// Exercises the production persistent-store/bootstrap path.
// No live services or production credentials are used by this test.
func TestAIAgentClientFreshHeartbeatSurvivesSnapshotReload(t *testing.T) {
	for _, reload := range []bool{false, true} {
		name := "without_reload"
		if reload {
			name = "with_reload"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			snapshots := &memoryAIAgentClientSnapshotStore{}
			principal := AuthorizationResult{PrincipalID: "diagnostic-user", WorkspaceID: "diagnostic-workspace"}
			store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
			if err != nil {
				t.Fatal(err)
			}
			enrollment, err := store.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, EnrollDeviceRequest{DisplayName: "Diagnostic Mac"})
			if err != nil {
				t.Fatal(err)
			}
			req := DeviceRuntimeSnapshotSyncRequest{
				DaemonID: enrollment.DeviceID, DeviceID: enrollment.DeviceID, Profile: "production", PID: 42,
				StartedAt: time.Now().Add(-10 * time.Minute),
				Runtimes:  []RuntimeSnapshotRecord{{RuntimeID: enrollment.DeviceID + ":codex", Kind: RuntimeKindCodex, Availability: RuntimeAvailabilityOnline, DetectionState: RuntimeDetectionStateDetected, Models: []RuntimeModelRecord{{ModelID: "diagnostic-model", Label: "Diagnostic", IsDefault: true}}}},
			}
			if _, err = store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, req); err != nil {
				t.Fatal(err)
			}
			staleAt := time.Now().Add(-2 * time.Minute)
			for i := range snapshots.snapshot.Devices {
				if snapshots.snapshot.Devices[i].DeviceID == enrollment.DeviceID {
					snapshots.snapshot.Devices[i].DaemonLastSeenAt = staleAt
					for j := range snapshots.snapshot.Devices[i].Runtimes {
						snapshots.snapshot.Devices[i].Runtimes[j].LastDetectedAt = staleAt
					}
				}
			}
			for i := range snapshots.snapshot.Daemons {
				if snapshots.snapshot.Daemons[i].DeviceID == enrollment.DeviceID {
					snapshots.snapshot.Daemons[i].LastSeenAt = staleAt
				}
			}
			snapshots.snapshot.SavedAt = staleAt
			store.snapshotLastSavedAt = staleAt
			store.snapshotLastReloadAt = time.Now()
			saves := snapshots.saves
			req.UptimeSeconds = 600
			hb, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, req)
			if err != nil {
				t.Fatal(err)
			}
			if snapshots.saves != saves {
				t.Fatal("setup failed: heartbeat unexpectedly persisted")
			}
			if hb.Device.Runtimes[0].Availability != RuntimeAvailabilityOnline {
				t.Fatal("setup failed: heartbeat was not online")
			}
			if reload {
				store.snapshotLastReloadAt = time.Now().Add(-2 * defaultAIAgentClientSnapshotReloadInterval)
			}
			view, err := store.BootstrapAIAgentClient(ctx, principal, ClientKindWeb)
			if err != nil {
				t.Fatal(err)
			}
			for _, d := range view.Devices {
				if d.DeviceID != enrollment.DeviceID {
					continue
				}
				for _, r := range d.Runtimes {
					t.Logf("heartbeat just accepted; reload=%v; availability=%s detection=%s models=%d", reload, r.Availability, r.DetectionState, len(r.Models))
					if r.Availability != RuntimeAvailabilityOnline {
						t.Fatalf("fresh heartbeat lost: runtime shown %s after durable snapshot reload (save=5m reload=1m stale=45s)", r.Availability)
					}
					return
				}
			}
			t.Fatal("diagnostic runtime missing from bootstrap")
		})
	}
}
