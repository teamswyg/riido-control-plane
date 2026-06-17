package riidoaiserver

import (
	"errors"
	"fmt"
	"strings"
)

var ErrAIAgentRuntimeAlreadyAssigned = errors.New("ai agent runtime is already assigned")

func (s *DevelopmentAIAgentClientStore) ensureRuntimeAssignableToAgentLocked(runtimeID, agentID string) error {
	runtimeID = strings.TrimSpace(runtimeID)
	agentID = strings.TrimSpace(agentID)
	if runtimeID == "" {
		return nil
	}
	for key, agent := range s.agents {
		existingAgentID := strings.TrimSpace(agent.AgentID)
		if existingAgentID == "" {
			existingAgentID = strings.TrimSpace(key)
		}
		if existingAgentID == agentID || strings.TrimSpace(agent.RuntimeID) != runtimeID {
			continue
		}
		return fmt.Errorf("%w: runtime_id %q is already assigned to agent %q", ErrAIAgentRuntimeAlreadyAssigned, runtimeID, existingAgentID)
	}
	return nil
}

func (s *DevelopmentAIAgentClientStore) refreshRuntimeAssignmentFlagLocked(runtimeID string) {
	runtimeID = strings.TrimSpace(runtimeID)
	if runtimeID == "" {
		return
	}
	markRuntimeHasAssignedAgentLocked(s.devices, runtimeID, s.runtimeHasAssignedAgentLocked(runtimeID))
}
