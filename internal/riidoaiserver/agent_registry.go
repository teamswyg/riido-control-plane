package riidoaiserver

import (
	"errors"
	"fmt"
	"strings"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

type AgentRuntimeBinding = assignmentcontract.AgentRuntimeBinding

type AgentRegistry interface {
	LookupAgent(agentID string) (AgentRuntimeBinding, bool)
}

type StaticAgentRegistry struct {
	byAgent map[string]AgentRuntimeBinding
}

func NewStaticAgentRegistry(bindings []AgentRuntimeBinding) (*StaticAgentRegistry, error) {
	byAgent := make(map[string]AgentRuntimeBinding, len(bindings))
	for i, binding := range bindings {
		binding = normalizeAgentRuntimeBinding(binding)
		if binding.AgentID == "" {
			return nil, fmt.Errorf("agent binding %d: agent_id is required", i)
		}
		if binding.DaemonID == "" {
			return nil, fmt.Errorf("agent binding %s: daemon_id is required", binding.AgentID)
		}
		if binding.RuntimeID == "" {
			return nil, fmt.Errorf("agent binding %s: runtime_id is required", binding.AgentID)
		}
		if binding.RuntimeProvider == "" {
			return nil, fmt.Errorf("agent binding %s: runtime_provider is required", binding.AgentID)
		}
		if _, exists := byAgent[binding.AgentID]; exists {
			return nil, fmt.Errorf("agent binding %s: duplicate agent_id", binding.AgentID)
		}
		byAgent[binding.AgentID] = binding
	}
	return &StaticAgentRegistry{byAgent: byAgent}, nil
}

func (r *StaticAgentRegistry) LookupAgent(agentID string) (AgentRuntimeBinding, bool) {
	if r == nil {
		return AgentRuntimeBinding{}, false
	}
	binding, ok := r.byAgent[strings.TrimSpace(agentID)]
	return binding, ok
}

func normalizeAgentRuntimeBinding(binding AgentRuntimeBinding) AgentRuntimeBinding {
	binding.AgentID = strings.TrimSpace(binding.AgentID)
	binding.DaemonID = strings.TrimSpace(binding.DaemonID)
	binding.DeviceID = strings.TrimSpace(binding.DeviceID)
	binding.RuntimeID = strings.TrimSpace(binding.RuntimeID)
	binding.RuntimeProvider = strings.TrimSpace(binding.RuntimeProvider)
	return binding
}

func validateAssignmentBinding(registry AgentRegistry, agentID, runtimeProvider string) error {
	if registry == nil {
		return nil
	}
	binding, ok := registry.LookupAgent(agentID)
	if !ok {
		return fmt.Errorf("agent %s is not registered", agentID)
	}
	if binding.RuntimeProvider != runtimeProvider {
		return fmt.Errorf("agent %s is bound to runtime_provider %s", agentID, binding.RuntimeProvider)
	}
	return nil
}

func validateDaemonBinding(registry AgentRegistry, agentID string, req PollRequest) error {
	if registry == nil {
		return nil
	}
	binding, ok := registry.LookupAgent(agentID)
	if !ok {
		return fmt.Errorf("agent %s is not registered", agentID)
	}
	req.DaemonID = strings.TrimSpace(req.DaemonID)
	req.DeviceID = strings.TrimSpace(req.DeviceID)
	req.RuntimeID = strings.TrimSpace(req.RuntimeID)
	if req.DaemonID == "" {
		return errors.New("daemon_id is required")
	}
	if req.RuntimeID == "" {
		return errors.New("runtime_id is required")
	}
	if req.DaemonID != binding.DaemonID {
		return fmt.Errorf("agent %s is bound to daemon_id %s", agentID, binding.DaemonID)
	}
	if binding.DeviceID != "" && req.DeviceID != binding.DeviceID {
		return fmt.Errorf("agent %s is bound to device_id %s", agentID, binding.DeviceID)
	}
	if req.RuntimeID != binding.RuntimeID {
		return fmt.Errorf("agent %s is bound to runtime_id %s", agentID, binding.RuntimeID)
	}
	return nil
}
