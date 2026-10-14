package riidoaiserver

import (
	"strings"
	"time"
)

func (s *DevelopmentAIAgentClientStore) agentRuntimeBindingLocked(
	agent AgentClientRecord,
) (AgentRuntimeBinding, bool) {
	binding, _, ok := s.agentRuntimeFactLocked(agent, time.Now().UTC())
	return binding, ok
}

func (s *DevelopmentAIAgentClientStore) agentRuntimeFactLocked(
	agent AgentClientRecord,
	now time.Time,
) (AgentRuntimeBinding, RuntimeRecord, bool) {
	agent.AgentID = strings.TrimSpace(agent.AgentID)
	agent.RuntimeID = strings.TrimSpace(agent.RuntimeID)
	if agent.AgentID == "" || agent.RuntimeID == "" {
		return AgentRuntimeBinding{}, RuntimeRecord{}, false
	}
	device, _, ok := s.deviceRuntimeByRuntimeIDLocked(agent.RuntimeID)
	if !ok {
		return AgentRuntimeBinding{}, RuntimeRecord{}, false
	}
	device = projectDeviceRuntimeLiveness(device, now)
	runtime, ok := runtimeByID(device.Runtimes, agent.RuntimeID)
	if !ok || !runtimeAvailableForBinding(runtime) {
		return AgentRuntimeBinding{}, RuntimeRecord{}, false
	}
	daemon, ok := s.daemonForRuntimeLocked(device.DeviceID, runtime)
	if !ok || strings.TrimSpace(daemon.DaemonID) == "" {
		return AgentRuntimeBinding{}, RuntimeRecord{}, false
	}
	daemon = projectDeviceDaemonLiveness(daemon, now)
	if daemon.Availability != DaemonAvailabilityOnline {
		return AgentRuntimeBinding{}, RuntimeRecord{}, false
	}
	provider := runtimeProviderForAIAgentRuntime(agent.RuntimeKind)
	if provider == "" {
		provider = runtimeProviderForAIAgentRuntime(runtime.Kind)
	}
	if provider == "" {
		return AgentRuntimeBinding{}, RuntimeRecord{}, false
	}
	return normalizeAgentRuntimeBinding(AgentRuntimeBinding{
		AgentID: agent.AgentID, DaemonID: daemon.DaemonID, DeviceID: device.DeviceID,
		RuntimeID: runtime.RuntimeID, RuntimeProvider: provider,
	}), copyRuntime(runtime), true
}
