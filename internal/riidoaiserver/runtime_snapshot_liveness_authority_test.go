package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestSnapshotLivenessPreservesDurableAuthority(t *testing.T) {
	for _, mode := range []string{"fresh", "deleted-runtime", "deleted-device", "revoked", "rotated", "new-session", "newer-durable", "offline"} {
		t.Run(mode, func(t *testing.T) {
			s := NewDevelopmentAIAgentClientStore()
			principal := AuthorizationResult{PrincipalID: "owner", WorkspaceID: "workspace"}
			enrolled, err := s.EnrollDeviceCredential(context.Background(), principal, "workspace", EnrollDeviceRequest{DisplayName: "fixture", MachineID: "fixture-machine"})
			if err != nil {
				t.Fatal(err)
			}
			req := DeviceRuntimeSnapshotSyncRequest{DeviceID: enrolled.DeviceID, DaemonID: enrolled.DeviceID, Profile: "production", PID: 3, StartedAt: time.Now().Add(-time.Hour), Runtimes: []RuntimeSnapshotRecord{{RuntimeID: enrolled.DeviceID + ":codex", Kind: RuntimeKindCodex, Availability: RuntimeAvailabilityOnline, DetectionState: RuntimeDetectionStateDetected}}}
			if _, err := s.SyncAIAgentDaemonRuntimeSnapshot(context.Background(), principal, req); err != nil {
				t.Fatal(err)
			}
			snapshot, err := s.snapshot(time.Now())
			if err != nil {
				t.Fatal(err)
			}
			old := time.Now().Add(-2 * time.Minute)
			deviceIndex := -1
			for i := range snapshot.Devices {
				if snapshot.Devices[i].DeviceID == enrolled.DeviceID {
					deviceIndex = i
					snapshot.Devices[i].DaemonLastSeenAt = old
				}
			}
			for i := range snapshot.Daemons {
				snapshot.Daemons[i].LastSeenAt = old
			}
			switch mode {
			case "deleted-runtime":
				snapshot.Devices[deviceIndex].Runtimes = nil
			case "deleted-device":
				snapshot.Devices = append(snapshot.Devices[:deviceIndex], snapshot.Devices[deviceIndex+1:]...)
			case "revoked":
				snapshot.DeviceCredentials = nil
			case "rotated":
				snapshot.DeviceCredentials[0].SecretHashSHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
			case "new-session":
				snapshot.Daemons[0].StartedAt = time.Now()
				snapshot.Daemons[0].PID = 4
			case "newer-durable":
				snapshot.Devices[deviceIndex].DaemonLastSeenAt = time.Now().Add(time.Minute)
				snapshot.Daemons[0].LastSeenAt = time.Now().Add(time.Minute)
			case "offline":
				snapshot.Daemons[0].Availability = DaemonAvailabilityOffline
			}
			if err := s.restoreSnapshotPreservingSubscribers(snapshot); err != nil {
				t.Fatal(err)
			}
			device, found := s.deviceByIDLocked(enrolled.DeviceID)
			if mode == "deleted-device" {
				if found {
					t.Fatal("resurrected deleted device")
				}
				return
			}
			if mode == "deleted-runtime" && len(device.Runtimes) > 0 {
				t.Fatal("resurrected deleted runtime")
			}
			if mode == "fresh" {
				if !device.DaemonLastSeenAt.After(old) {
					t.Fatal("lost local heartbeat")
				}
				return
			}
			expected := snapshot.Devices[deviceIndex].DaemonLastSeenAt
			if !device.DaemonLastSeenAt.Equal(expected) {
				t.Fatalf("overrode durable authority: got %s want %s", device.DaemonLastSeenAt, expected)
			}
		})
	}
}
