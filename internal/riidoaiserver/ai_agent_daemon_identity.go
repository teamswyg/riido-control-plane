package riidoaiserver

import (
	"strings"
)

const daemonIdentitySeparator = "\x00"

func daemonStorageKey(daemon DeviceDaemonRecord) string {
	return daemonStorageKeyFor(daemon.DeviceID, daemon.Profile, daemon.DaemonID)
}

func daemonStorageKeyFor(deviceID, profile, daemonID string) string {
	deviceID = strings.TrimSpace(deviceID)
	scope := strings.TrimSpace(profile)
	if scope == "" {
		scope = strings.TrimSpace(daemonID)
	}
	if scope == "" {
		return deviceID
	}
	return deviceID + daemonIdentitySeparator + scope
}

func daemonBelongsToDevice(daemon DeviceDaemonRecord, deviceID string) bool {
	return strings.TrimSpace(daemon.DeviceID) == strings.TrimSpace(deviceID)
}

func sameDaemonIdentity(a, b DeviceDaemonRecord) bool {
	if !daemonBelongsToDevice(a, b.DeviceID) {
		return false
	}
	if strings.TrimSpace(a.Profile) != "" && strings.TrimSpace(b.Profile) != "" {
		return strings.TrimSpace(a.Profile) == strings.TrimSpace(b.Profile)
	}
	if strings.TrimSpace(a.DaemonID) != "" && strings.TrimSpace(b.DaemonID) != "" {
		return strings.TrimSpace(a.DaemonID) == strings.TrimSpace(b.DaemonID)
	}
	return false
}

func daemonBetterForDevice(candidate, current DeviceDaemonRecord) bool {
	if current.DeviceID == "" {
		return true
	}
	if candidate.LastSeenAt.After(current.LastSeenAt) {
		return true
	}
	if candidate.LastSeenAt.Equal(current.LastSeenAt) {
		return candidate.StartedAt.After(current.StartedAt)
	}
	return false
}

func normalizeDaemonRuntimeScope(profile, daemonID string) (string, string) {
	return strings.TrimSpace(profile), strings.TrimSpace(daemonID)
}
