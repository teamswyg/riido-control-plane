package riidoaiserver

import (
	"context"
	"testing"
)

func enrollTestDevice(
	t *testing.T,
	ctx context.Context,
	store *DevelopmentAIAgentClientStore,
	principal AuthorizationResult,
	workspaceID string,
	machine string,
	displayName string,
) EnrollDeviceResponse {
	t.Helper()
	enroll, err := store.EnrollDeviceCredential(ctx, principal, workspaceID,
		EnrollDeviceRequest{MachineID: machine, DisplayName: displayName})
	if err != nil {
		t.Fatalf("enroll %s: %v", workspaceID, err)
	}
	return enroll
}

func syncTestRuntime(
	t *testing.T,
	ctx context.Context,
	store *DevelopmentAIAgentClientStore,
	principal AuthorizationResult,
	daemonID string,
	deviceID string,
	runtimeID string,
) {
	t.Helper()
	_, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal,
		DeviceRuntimeSnapshotSyncRequest{
			DaemonID: daemonID,
			DeviceID: deviceID,
			Runtimes: []RuntimeSnapshotRecord{{
				RuntimeID:      runtimeID,
				Kind:           RuntimeKindClaudeCode,
				Availability:   RuntimeAvailabilityOnline,
				DetectionState: RuntimeDetectionStateDetected,
			}},
		})
	if err != nil {
		t.Fatalf("sync runtime snapshot: %v", err)
	}
}

func requireDevice(t *testing.T, devices []DeviceRecord, deviceID string) DeviceRecord {
	t.Helper()
	for i := range devices {
		if devices[i].DeviceID == deviceID {
			return devices[i]
		}
	}
	t.Fatalf("device %q not found in %+v", deviceID, devices)
	return DeviceRecord{}
}

func containsDeviceID(devices []DeviceRecord, deviceID string) bool {
	for _, d := range devices {
		if d.DeviceID == deviceID {
			return true
		}
	}
	return false
}
