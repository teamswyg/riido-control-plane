package riidoaiserver

import (
	"errors"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) applyAgentRuntimePatchLocked(agent *AgentClientRecord, principal AuthorizationResult, previousRuntimeID string, req UpdateAgentConfigurationRequest) error {
	nextRuntimeID := agent.RuntimeID
	if strings.TrimSpace(req.RuntimeID) != "" {
		nextRuntimeID = strings.TrimSpace(req.RuntimeID)
	}
	if strings.TrimSpace(req.RuntimeID) == "" && req.ModelID == nil {
		return nil
	}
	runtimeKind, runtimeModel, ok := runtimeSelectionFromDevices(s.visibleDevicesLocked(principal), nextRuntimeID, req.ModelID)
	if !ok {
		return errors.New("runtime_id or model_id is not available")
	}
	agent.RuntimeID = nextRuntimeID
	agent.RuntimeKind = runtimeKind
	agent.ModelID = runtimeModel.ModelID
	agent.ModelLabel = runtimeModel.Label
	return nil
}
