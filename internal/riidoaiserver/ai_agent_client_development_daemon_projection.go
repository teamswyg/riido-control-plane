package riidoaiserver

import (
	"slices"
	"time"
)

func deviceDaemonFromDeviceReadModel(device DeviceRecord) DeviceDaemonRecord {
	daemon := DeviceDaemonRecord{
		DeviceID:          device.DeviceID,
		OwnerPrincipalID:  device.OwnerPrincipalID,
		DeviceDisplayName: device.DisplayName,
		LastSeenAt:        device.DaemonLastSeenAt,
		Availability:      DaemonAvailabilityOffline,
		ControlState:      DaemonControlStateIdle,
		SupportedActions:  []DaemonControlAction{DaemonControlActionStart},
	}
	for _, runtime := range device.Runtimes {
		if runtime.Availability != RuntimeAvailabilityOnline {
			continue
		}
		daemon.Availability = DaemonAvailabilityOnline
		daemon.SupportedActions = []DaemonControlAction{DaemonControlActionRestart, DaemonControlActionStop}
		return daemon
	}
	return daemon
}

func projectDeviceDaemonLiveness(daemon DeviceDaemonRecord, now time.Time) DeviceDaemonRecord {
	daemon = copyDeviceDaemon(daemon)
	if isDevelopmentSeedDevice(DeviceRecord{DeviceID: daemon.DeviceID}) {
		return daemon
	}
	if !deviceRuntimeSnapshotStale(daemon.LastSeenAt, now) {
		return daemon
	}
	daemon.Availability = DaemonAvailabilityOffline
	daemon.ControlState = DaemonControlStateIdle
	daemon.SupportedActions = []DaemonControlAction{DaemonControlActionStart}
	daemon.PID = 0
	daemon.UptimeSeconds = 0
	return daemon
}

func daemonSupportsAction(daemon DeviceDaemonRecord, action DaemonControlAction) bool {
	return slices.Contains(daemon.SupportedActions, action)
}
