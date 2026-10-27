package riidoaiserver

import (
	"strings"
	"time"
)

// visibleAgentDaemonLastSeenLocked reuses the assignable-list snapshot. It
// returns only visible identities and timestamps, with no per-agent store calls.
func (s *DevelopmentAIAgentClientStore) visibleAgentDaemonLastSeenLocked(agents []AgentClientRecord) map[string]time.Time {
	wanted := make(map[string]bool, len(agents))
	for _, agent := range agents {
		if id := strings.TrimSpace(agent.RuntimeID); id != "" {
			wanted[id] = true
		}
	}
	if len(wanted) == 0 {
		return nil
	}
	byRuntime := make(map[string]time.Time, len(wanted))
	for _, device := range s.devices {
		for _, runtime := range device.Runtimes {
			if !wanted[runtime.RuntimeID] {
				continue
			}
			// Match the first runtime record, like deviceRuntimeByRuntimeIDLocked.
			delete(wanted, runtime.RuntimeID)
			lastSeen := s.directoryRuntimeHeartbeatLocked(device, runtime)
			if !lastSeen.IsZero() {
				byRuntime[runtime.RuntimeID] = lastSeen.UTC()
			}
		}
	}
	out := make(map[string]time.Time, len(agents))
	for _, agent := range agents {
		if lastSeen, ok := byRuntime[strings.TrimSpace(agent.RuntimeID)]; ok {
			out[agent.AgentID] = lastSeen
		}
	}
	return out
}

func (s *DevelopmentAIAgentClientStore) directoryRuntimeHeartbeatLocked(device DeviceRecord, runtime RuntimeRecord) time.Time {
	profile, daemonID := normalizeDaemonRuntimeScope(runtime.DaemonProfile, runtime.DaemonID)
	if profile == "" && daemonID == "" {
		return device.DaemonLastSeenAt
	}
	wanted := DeviceDaemonRecord{DeviceID: device.DeviceID, Profile: profile, DaemonID: daemonID}
	if daemon, ok := s.daemons[daemonStorageKey(wanted)]; ok && sameDaemonIdentity(daemon, wanted) {
		return daemon.LastSeenAt
	}
	// Another daemon on the same device cannot establish this runtime's heartbeat.
	return time.Time{}
}
