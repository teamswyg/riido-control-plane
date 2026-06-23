package riidoaiserver

import "strings"

func (s *DevelopmentAIAgentClientStore) refreshRuntimeAssignmentFlagLocked(runtimeID string) {
	runtimeID = strings.TrimSpace(runtimeID)
	if runtimeID == "" {
		return
	}
	markRuntimeHasAssignedAgentLocked(s.devices, runtimeID, s.runtimeHasAssignedAgentLocked(runtimeID))
}
