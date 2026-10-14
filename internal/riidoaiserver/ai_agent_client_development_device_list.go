package riidoaiserver

import "time"

func (s *DevelopmentAIAgentClientStore) visibleDevicesLocked(principal AuthorizationResult) []DeviceRecord {
	now := time.Now().UTC()
	visibleRuntimeIDs := map[string]struct{}{}
	for _, agent := range s.agents {
		if !s.aiAgentVisibleTo(principal, agent) || agent.RuntimeID == "" {
			continue
		}
		visibleRuntimeIDs[agent.RuntimeID] = struct{}{}
	}
	out := make([]DeviceRecord, 0, len(s.devices))
	for _, device := range s.devices {
		if deviceHiddenSeedForPrincipal(device, principal) {
			continue
		}
		device = projectDeviceRuntimeLiveness(device, now)
		device.Runtimes = dedupeRuntimesByKindForDisplay(device.Runtimes)
		device = s.projectDeviceClientStatusLocked(principal, device, now)
		if device.OwnerPrincipalID == principal.PrincipalID {
			out = append(out, copyDevice(device))
			continue
		}
		if deviceConnectedToWorkspace(device, principal.WorkspaceID) {
			out = append(out, copyDevice(device))
			continue
		}
		if s.deviceWorkspaceVisibleByAdminLocked(principal, device.DeviceID) {
			out = append(out, copyDevice(device))
			continue
		}
		filtered := device
		filtered.Runtimes = nil
		for _, runtime := range device.Runtimes {
			if _, ok := visibleRuntimeIDs[runtime.RuntimeID]; ok {
				filtered.Runtimes = append(filtered.Runtimes, copyRuntime(runtime))
			}
		}
		if len(filtered.Runtimes) > 0 {
			out = append(out, filtered)
		}
	}
	return out
}
