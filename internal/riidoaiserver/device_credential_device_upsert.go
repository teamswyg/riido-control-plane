package riidoaiserver

func (s *DevelopmentAIAgentClientStore) upsertEnrolledDeviceLocked(device DeviceRecord) {
	for i := range s.devices {
		if s.devices[i].DeviceID == device.DeviceID {
			merged := copyDevice(s.devices[i])
			merged.OwnerPrincipalID = device.OwnerPrincipalID
			if device.DisplayName != "" {
				merged.DisplayName = device.DisplayName
			}
			if device.DesktopAppVersion != "" {
				merged.DesktopAppVersion = device.DesktopAppVersion
			}
			if !device.DaemonLastSeenAt.IsZero() {
				merged.DaemonLastSeenAt = device.DaemonLastSeenAt
			}
			for _, ws := range device.ConnectedWorkspaceIDs {
				merged.ConnectedWorkspaceIDs = addConnectedWorkspace(merged.ConnectedWorkspaceIDs, ws)
			}
			s.devices[i] = merged
			return
		}
	}
	s.devices = append(s.devices, copyDevice(device))
}
