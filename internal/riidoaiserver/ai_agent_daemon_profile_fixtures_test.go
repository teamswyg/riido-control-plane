package riidoaiserver

import "testing"

func profileRuntimeSnapshot(deviceID, profile string, kind RuntimeKind) DeviceRuntimeSnapshotSyncRequest {
	return DeviceRuntimeSnapshotSyncRequest{
		DaemonID: "daemon-" + profile,
		DeviceID: deviceID,
		Profile:  profile,
		Runtimes: []RuntimeSnapshotRecord{{
			RuntimeID:      deviceID + ":" + profile + ":" + string(kind),
			Kind:           kind,
			Availability:   RuntimeAvailabilityOnline,
			DetectionState: RuntimeDetectionStateDetected,
		}},
	}
}

func profileRuntime(deviceID, profile string, kind RuntimeKind) RuntimeRecord {
	return RuntimeRecord{
		RuntimeID:      deviceID + ":" + profile + ":" + string(kind),
		DeviceID:       deviceID,
		DaemonID:       "daemon-" + profile,
		DaemonProfile:  profile,
		Kind:           kind,
		Availability:   RuntimeAvailabilityOnline,
		DetectionState: RuntimeDetectionStateDetected,
	}
}

func profileDaemon(deviceID, profile string) DeviceDaemonRecord {
	return DeviceDaemonRecord{
		DeviceID:         deviceID,
		DaemonID:         "daemon-" + profile,
		Profile:          profile,
		Availability:     DaemonAvailabilityOnline,
		ControlState:     DaemonControlStateIdle,
		SupportedActions: []DaemonControlAction{DaemonControlActionRestart, DaemonControlActionStop},
	}
}

func assertOnlyProfileRuntime(t *testing.T, store *DevelopmentAIAgentClientStore, deviceID, profile string) {
	t.Helper()
	store.mu.Lock()
	defer store.mu.Unlock()
	for _, device := range store.devices {
		if device.DeviceID != deviceID {
			continue
		}
		assertDeviceRuntimesProfile(t, device, profile)
		return
	}
	t.Fatalf("device %s not found", deviceID)
}

func assertDeviceRuntimesProfile(t *testing.T, device DeviceRecord, profile string) {
	t.Helper()
	for _, runtime := range device.Runtimes {
		if runtime.DaemonProfile != profile {
			t.Fatalf("runtime %s profile=%q, want %q",
				runtime.RuntimeID, runtime.DaemonProfile, profile)
		}
	}
}
