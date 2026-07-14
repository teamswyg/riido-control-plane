package riidoaiserver

import "time"

func (s *DevelopmentAIAgentClientStore) projectDeviceClientStatusLocked(
	principal AuthorizationResult,
	device DeviceRecord,
	now time.Time,
) DeviceRecord {
	device.IsOwnedByViewer = device.OwnerPrincipalID == principal.PrincipalID
	status := DeviceClientStatus{
		DesktopRegistered:    true,
		DesktopAppVersion:    device.DesktopAppVersion,
		DaemonAvailability:   DaemonAvailabilityOffline,
		MinimumDaemonVersion: s.daemonClientPolicy.MinimumVersion,
		LatestDaemonVersion:  s.daemonClientPolicy.LatestVersion,
		AgentCapability:      DeviceAgentCapabilityOffline,
		DownloadURL:          s.daemonClientPolicy.DownloadURL,
	}
	if daemon, ok := s.preferredDaemonForDeviceLocked(device.DeviceID); ok {
		daemon = projectDeviceDaemonLiveness(daemon, now)
		status.DaemonAvailability = daemon.Availability
		status.DaemonVersion = daemon.AppVersion
		classifyDeviceAgentCapability(&status)
	}
	device.ClientStatus = &status
	return device
}

func classifyDeviceAgentCapability(status *DeviceClientStatus) {
	if status.DaemonAvailability != DaemonAvailabilityOnline {
		return
	}
	comparison, ok := compareDaemonVersions(status.DaemonVersion, status.MinimumDaemonVersion)
	if !ok {
		status.AgentCapability = DeviceAgentCapabilityVersionUnknown
		return
	}
	if comparison < 0 {
		status.AgentCapability = DeviceAgentCapabilityUpdateRequired
		status.UpdateRequired = true
		return
	}
	status.AgentCapability = DeviceAgentCapabilityReady
	status.AgentSupported = true
}
