package riidoaiserver

import "strings"

func (s *DevelopmentAIAgentClientStore) daemonByIdentityLocked(deviceID, profile, daemonID string) (DeviceDaemonRecord, bool) {
	key := daemonStorageKeyFor(deviceID, profile, daemonID)
	if daemon, ok := s.daemons[key]; ok {
		return copyDeviceDaemon(daemon), true
	}
	if daemon, ok := s.daemons[strings.TrimSpace(deviceID)]; ok {
		return copyDeviceDaemon(daemon), true
	}
	return DeviceDaemonRecord{}, false
}

func (s *DevelopmentAIAgentClientStore) putDaemonLocked(daemon DeviceDaemonRecord) {
	key := daemonStorageKey(daemon)
	legacyKey := strings.TrimSpace(daemon.DeviceID)
	if key != legacyKey {
		if legacy, ok := s.daemons[legacyKey]; ok && sameDaemonIdentity(legacy, daemon) {
			delete(s.daemons, legacyKey)
		}
	}
	s.daemons[key] = daemon
}

func (s *DevelopmentAIAgentClientStore) preferredDaemonForDeviceLocked(deviceID string) (DeviceDaemonRecord, bool) {
	var best DeviceDaemonRecord
	for _, daemon := range s.daemons {
		if !daemonBelongsToDevice(daemon, deviceID) {
			continue
		}
		if daemonBetterForDevice(daemon, best) {
			best = daemon
		}
	}
	if best.DeviceID == "" {
		return DeviceDaemonRecord{}, false
	}
	return copyDeviceDaemon(best), true
}

func (s *DevelopmentAIAgentClientStore) daemonForRuntimeLocked(deviceID string, runtime RuntimeRecord) (DeviceDaemonRecord, bool) {
	if daemon, ok := s.daemonByIdentityLocked(deviceID, runtime.DaemonProfile, runtime.DaemonID); ok {
		return daemon, true
	}
	return s.preferredDaemonForDeviceLocked(deviceID)
}
