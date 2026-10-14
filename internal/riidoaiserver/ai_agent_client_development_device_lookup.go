package riidoaiserver

import "time"

func (s *DevelopmentAIAgentClientStore) deviceByRuntimeIDLocked(runtimeID string) (DeviceRecord, bool) {
	for _, device := range s.devices {
		for _, runtime := range device.Runtimes {
			if runtime.RuntimeID == runtimeID {
				return copyDevice(device), true
			}
		}
	}
	return DeviceRecord{}, false
}

func (s *DevelopmentAIAgentClientStore) deviceByIDLocked(deviceID string) (DeviceRecord, bool) {
	for _, device := range s.devices {
		if device.DeviceID == deviceID {
			return copyDevice(device), true
		}
	}
	return DeviceRecord{}, false
}

func (s *DevelopmentAIAgentClientStore) markDeviceRuntimesOfflineLocked(deviceID string, observedAt time.Time) {
	for deviceIndex := range s.devices {
		if s.devices[deviceIndex].DeviceID != deviceID {
			continue
		}
		s.devices[deviceIndex].DaemonLastSeenAt = observedAt
		for runtimeIndex := range s.devices[deviceIndex].Runtimes {
			s.devices[deviceIndex].Runtimes[runtimeIndex].Availability = RuntimeAvailabilityOffline
			s.devices[deviceIndex].Runtimes[runtimeIndex].DetectionState = RuntimeDetectionStateMissing
			s.devices[deviceIndex].Runtimes[runtimeIndex].LastDetectedAt = observedAt
		}
		return
	}
}
