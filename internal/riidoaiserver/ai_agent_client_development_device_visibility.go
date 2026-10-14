package riidoaiserver

import "time"

func (s *DevelopmentAIAgentClientStore) visibleDeviceRecordLocked(principal AuthorizationResult, device DeviceRecord) (DeviceRecord, bool) {
	if deviceHiddenSeedForPrincipal(device, principal) {
		return DeviceRecord{}, false
	}
	device = projectDeviceRuntimeLiveness(device, time.Now().UTC())
	device.Runtimes = dedupeRuntimesByKindForDisplay(device.Runtimes)
	if device.OwnerPrincipalID == principal.PrincipalID {
		return copyDevice(device), true
	}
	if s.deviceWorkspaceVisibleByAdminLocked(principal, device.DeviceID) {
		return copyDevice(device), true
	}
	filtered := device
	filtered.Runtimes = nil
	for _, runtime := range device.Runtimes {
		if s.runtimeVisibleThroughAgentLocked(principal, runtime.RuntimeID) {
			filtered.Runtimes = append(filtered.Runtimes, copyRuntime(runtime))
		}
	}
	if len(filtered.Runtimes) == 0 {
		return DeviceRecord{}, false
	}
	return filtered, true
}

func (s *DevelopmentAIAgentClientStore) deviceVisibleToPrincipalLocked(principal AuthorizationResult, device DeviceRecord) bool {
	if deviceHiddenSeedForPrincipal(device, principal) {
		return false
	}
	if device.OwnerPrincipalID == principal.PrincipalID {
		return true
	}
	if deviceConnectedToWorkspace(device, principal.WorkspaceID) {
		return true
	}
	if s.deviceWorkspaceVisibleByAdminLocked(principal, device.DeviceID) {
		return true
	}
	for _, runtime := range device.Runtimes {
		if s.runtimeVisibleThroughAgentLocked(principal, runtime.RuntimeID) {
			return true
		}
	}
	return false
}
