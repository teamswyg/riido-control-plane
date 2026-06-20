package riidoaiserver

import "time"

func developmentPrimaryDaemon(device DeviceRecord, now time.Time) DeviceDaemonRecord {
	return DeviceDaemonRecord{
		DeviceID:          device.DeviceID,
		OwnerPrincipalID:  device.OwnerPrincipalID,
		DeviceDisplayName: device.DisplayName,
		DaemonID:          "daemon-dev-macbook",
		Profile:           "desktop-api.riido.ai",
		PID:               5111,
		UptimeSeconds:     74 * 60,
		StartedAt:         now.Add(-74 * time.Minute),
		LastSeenAt:        now,
		Availability:      DaemonAvailabilityOnline,
		ControlState:      DaemonControlStateIdle,
		SupportedActions:  []DaemonControlAction{DaemonControlActionRestart, DaemonControlActionStop},
	}
}

func developmentSharedDaemon(device DeviceRecord, now time.Time) DeviceDaemonRecord {
	return DeviceDaemonRecord{
		DeviceID:          device.DeviceID,
		OwnerPrincipalID:  device.OwnerPrincipalID,
		DeviceDisplayName: device.DisplayName,
		DaemonID:          "daemon-shared-studio",
		Profile:           "desktop-api.riido.ai",
		PID:               6111,
		UptimeSeconds:     42 * 60,
		StartedAt:         now.Add(-42 * time.Minute),
		LastSeenAt:        now,
		Availability:      DaemonAvailabilityOnline,
		ControlState:      DaemonControlStateIdle,
		SupportedActions:  []DaemonControlAction{DaemonControlActionRestart, DaemonControlActionStop},
	}
}
