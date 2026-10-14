package riidoaiserver

import "strings"

func (s *DevelopmentAIAgentClientStore) deviceWorkspaceVisibleByAdminLocked(principal AuthorizationResult, deviceID string) bool {
	if !aiAgentIsAdmin(principal) {
		return false
	}
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return false
	}
	for _, record := range s.deviceCredentials {
		if record.deviceID == deviceID && record.workspaceID == s.workspaceScope(principal) {
			return true
		}
	}
	return isDevelopmentFixturePrincipal(principal) && isDevelopmentSeedDevice(DeviceRecord{DeviceID: deviceID})
}

func (s *DevelopmentAIAgentClientStore) runtimeVisibleThroughAgentLocked(principal AuthorizationResult, runtimeID string) bool {
	for _, agent := range s.agents {
		if agent.RuntimeID == runtimeID && s.aiAgentVisibleTo(principal, agent) {
			return true
		}
	}
	return false
}

func deviceHiddenSeedForPrincipal(device DeviceRecord, principal AuthorizationResult) bool {
	return isDevelopmentSeedDevice(device) && !isDevelopmentFixturePrincipal(principal)
}

func isDevelopmentSeedDevice(device DeviceRecord) bool {
	switch strings.TrimSpace(device.DeviceID) {
	case "device-dev-macbook", "device-shared-studio":
		return true
	default:
		return false
	}
}

func isDevelopmentFixturePrincipal(principal AuthorizationResult) bool {
	switch strings.TrimSpace(principal.PrincipalID) {
	case "user-1", "user-2", "admin-1", "admin-user":
		return true
	default:
		return false
	}
}
