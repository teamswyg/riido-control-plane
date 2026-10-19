package riidoaiserver

import "reflect"

func sameEnrollmentCredential(left, right AIAgentClientSnapshot, deviceID string) bool {
	l, lok := snapshotDeviceCredential(left, deviceID)
	r, rok := snapshotDeviceCredential(right, deviceID)
	return lok && rok && reflect.DeepEqual(l, r)
}

func snapshotDeviceCredential(snapshot AIAgentClientSnapshot, deviceID string) (AIAgentClientDeviceCredentialSnapshot, bool) {
	for _, credential := range snapshot.DeviceCredentials {
		if credential.DeviceID == deviceID {
			return credential, true
		}
	}
	return AIAgentClientDeviceCredentialSnapshot{}, false
}

// The caller holds operationMu. Publish only enrollment-owned fields; an
// unrelated current heartbeat/runtime update must never be restored from trial.
func (s *PersistentAIAgentClientStore) publishDeviceEnrollment(trial *DevelopmentAIAgentClientStore, deviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deviceCredentials[deviceID] = trial.deviceCredentials[deviceID]
	for _, device := range trial.devices {
		if device.DeviceID != deviceID {
			continue
		}
		s.upsertEnrolledDeviceLocked(device)
	}
	if s.deviceConnectionGrants[deviceID] == nil {
		s.deviceConnectionGrants[deviceID] = map[string]DeviceConnectionGrant{}
	}
	for key, grant := range trial.deviceConnectionGrants[deviceID] {
		old := s.deviceConnectionGrants[deviceID][key]
		if old.LastSeenAt.After(grant.LastSeenAt) {
			grant.LastSeenAt = old.LastSeenAt
		}
		s.deviceConnectionGrants[deviceID][key] = grant
	}
	if trial.nextDeviceCredentialSeq > s.nextDeviceCredentialSeq {
		s.nextDeviceCredentialSeq = trial.nextDeviceCredentialSeq
	}
}
