package riidoaiserver

import "time"

func (s *DevelopmentAIAgentClientStore) deviceDaemonForOwnerLocked(principal AuthorizationResult, deviceID string) (DeviceDaemonRecord, bool) {
	if daemon, ok := s.preferredDaemonForDeviceLocked(deviceID); ok && daemon.OwnerPrincipalID == principal.PrincipalID {
		return projectDeviceDaemonLiveness(daemon, time.Now().UTC()), true
	}
	for _, device := range s.devices {
		if device.DeviceID != deviceID || device.OwnerPrincipalID != principal.PrincipalID {
			continue
		}
		daemon, ok := s.preferredDaemonForDeviceLocked(deviceID)
		if !ok {
			return projectDeviceDaemonLiveness(deviceDaemonFromDeviceReadModel(device), time.Now().UTC()), true
		}
		return projectDeviceDaemonLiveness(daemon, time.Now().UTC()), true
	}
	for _, agent := range s.agents {
		if agent.OwnerPrincipalID != principal.PrincipalID {
			continue
		}
		device, ok := s.deviceByRuntimeIDLocked(agent.RuntimeID)
		if !ok || device.DeviceID != deviceID {
			continue
		}
		daemon, ok := s.preferredDaemonForDeviceLocked(device.DeviceID)
		if !ok {
			return projectDeviceDaemonLiveness(deviceDaemonFromDeviceReadModel(device), time.Now().UTC()), true
		}
		return projectDeviceDaemonLiveness(daemon, time.Now().UTC()), true
	}
	return DeviceDaemonRecord{}, false
}
