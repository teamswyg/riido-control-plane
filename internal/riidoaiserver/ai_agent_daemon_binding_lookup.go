package riidoaiserver

import "strings"

// LookupDaemonBinding authenticates a daemon poll against the durable current
// binding. It intentionally does not project heartbeat age into availability:
// the authenticated poll is fresh liveness evidence from that daemon.
func (s *DevelopmentAIAgentClientStore) LookupDaemonBinding(agentID string) (AgentRuntimeBinding, bool) {
	if s == nil {
		return AgentRuntimeBinding{}, false
	}
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return AgentRuntimeBinding{}, false
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.agents[agentID]
	if !ok {
		return AgentRuntimeBinding{}, false
	}
	return s.daemonBindingLocked(agent)
}

func (s *DevelopmentAIAgentClientStore) daemonBindingLocked(agent AgentClientRecord) (AgentRuntimeBinding, bool) {
	agent.AgentID = strings.TrimSpace(agent.AgentID)
	agent.RuntimeID = strings.TrimSpace(agent.RuntimeID)
	if agent.AgentID == "" || agent.RuntimeID == "" {
		return AgentRuntimeBinding{}, false
	}
	device, _, ok := s.deviceRuntimeByRuntimeIDLocked(agent.RuntimeID)
	if !ok {
		return AgentRuntimeBinding{}, false
	}
	runtime, ok := runtimeByID(device.Runtimes, agent.RuntimeID)
	if !ok || !runtimeAvailableForBinding(runtime) {
		return AgentRuntimeBinding{}, false
	}
	daemon, ok := s.daemonForRuntimeLocked(device.DeviceID, runtime)
	if !ok || strings.TrimSpace(daemon.DaemonID) == "" || daemon.Availability != DaemonAvailabilityOnline {
		return AgentRuntimeBinding{}, false
	}
	provider := runtimeProviderForAIAgentRuntime(agent.RuntimeKind)
	if provider == "" {
		provider = runtimeProviderForAIAgentRuntime(runtime.Kind)
	}
	if provider == "" {
		return AgentRuntimeBinding{}, false
	}
	return normalizeAgentRuntimeBinding(AgentRuntimeBinding{
		AgentID: agent.AgentID, DaemonID: daemon.DaemonID, DeviceID: device.DeviceID,
		RuntimeID: runtime.RuntimeID, RuntimeProvider: provider,
	}), true
}
