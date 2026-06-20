package riidoaiserver

import "time"

func (s *DevelopmentAIAgentClientStore) daemonVisibleToPrincipalLocked(principal AuthorizationResult, daemon DeviceDaemonRecord) bool {
	if daemon.OwnerPrincipalID == principal.PrincipalID {
		return true
	}
	for _, device := range s.devices {
		if device.DeviceID != daemon.DeviceID {
			continue
		}
		return s.deviceVisibleToPrincipalLocked(principal, device)
	}
	return false
}

func (s *DevelopmentAIAgentClientStore) deviceDaemonForAgentAccessLocked(principal AuthorizationResult, agentID string) (AgentClientRecord, DeviceDaemonRecord, bool) {
	agent, ok := s.visibleAgent(principal, agentID)
	if !ok {
		return AgentClientRecord{}, DeviceDaemonRecord{}, false
	}
	device, ok := s.deviceByRuntimeIDLocked(agent.RuntimeID)
	if !ok {
		return AgentClientRecord{}, DeviceDaemonRecord{}, false
	}
	daemon, ok := s.daemons[device.DeviceID]
	if !ok {
		return AgentClientRecord{}, DeviceDaemonRecord{}, false
	}
	return agent, projectDeviceDaemonLiveness(daemon, time.Now().UTC()), true
}
