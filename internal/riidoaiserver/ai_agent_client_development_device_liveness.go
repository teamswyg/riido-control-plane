package riidoaiserver

import "time"

func projectDeviceRuntimeLiveness(device DeviceRecord, now time.Time) DeviceRecord {
	device = copyDevice(device)
	if isDevelopmentSeedDevice(device) {
		return device
	}
	if !deviceRuntimeSnapshotStale(device.DaemonLastSeenAt, now) {
		return device
	}
	for i := range device.Runtimes {
		device.Runtimes[i].Availability = RuntimeAvailabilityOffline
		device.Runtimes[i].DetectionState = RuntimeDetectionStateMissing
	}
	return device
}

func deviceRuntimeSnapshotStale(lastSeenAt, now time.Time) bool {
	if lastSeenAt.IsZero() {
		return false
	}
	return now.UTC().Sub(lastSeenAt.UTC()) > AIAgentDeviceRuntimeSnapshotStaleAfter
}
