package riidoaiserver

// Device ownership and installed desktop version belong to enrollment. Preserve
// newer committed rows while allowing a fresh daemon heartbeat to update its
// name/runtime data. Workspace connections are monotonic historical facts.
func mergeEnrollmentDevices(requested *AIAgentClientSnapshot, committed AIAgentClientSnapshot, deviceID string, enrolling bool) {
	positions := make(map[string]int, len(requested.Devices))
	for i, device := range requested.Devices {
		positions[device.DeviceID] = i
	}
	for _, old := range committed.Devices {
		i, exists := positions[old.DeviceID]
		if !exists {
			requested.Devices = append(requested.Devices, copyDevice(old))
			continue
		}
		device := &requested.Devices[i]
		if !enrolling || device.DeviceID != deviceID {
			device.OwnerPrincipalID, device.DesktopAppVersion = old.OwnerPrincipalID, old.DesktopAppVersion
			if !device.DaemonLastSeenAt.After(old.DaemonLastSeenAt) {
				device.DisplayName = old.DisplayName
			}
		}
		if old.DaemonLastSeenAt.After(device.DaemonLastSeenAt) {
			device.DaemonLastSeenAt, device.Runtimes = old.DaemonLastSeenAt, old.Runtimes
		}
		for _, workspace := range old.ConnectedWorkspaceIDs {
			device.ConnectedWorkspaceIDs = addConnectedWorkspace(device.ConnectedWorkspaceIDs, workspace)
		}
	}
}
