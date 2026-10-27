package riidoaiserver

import (
	"encoding/json"
	"testing"
	"time"
)

func directoryWireEntry(t *testing.T, body, agentID string) map[string]any {
	t.Helper()
	var reply struct{ Data map[string]any }
	if err := json.Unmarshal([]byte(body), &reply); err != nil {
		t.Fatal(err)
	}
	if entry, ok := reply.Data["agent"].(map[string]any); ok {
		return entry
	}
	for _, candidate := range reply.Data["agents"].([]any) {
		if entry := candidate.(map[string]any); entry["agent_id"] == agentID {
			return entry
		}
	}
	t.Fatal("visible agent not found")
	return nil
}

func setDirectoryHeartbeat(t *testing.T, store *DevelopmentAIAgentClientStore, agentID string, heartbeat time.Time, explicit bool) {
	t.Helper()
	store.mu.Lock()
	defer store.mu.Unlock()
	agent := store.agents[agentID]
	agent.UpdatedAt = time.Now().UTC().Add(3 * time.Hour)
	store.agents[agentID] = agent
	for i := range store.devices {
		device := &store.devices[i]
		for j := range device.Runtimes {
			runtime := &device.Runtimes[j]
			if runtime.RuntimeID != agent.RuntimeID {
				continue
			}
			runtime.Availability = RuntimeAvailabilityOffline
			runtime.LastDetectedAt = time.Now().UTC().Add(2 * time.Hour)
			device.DaemonLastSeenAt = time.Now().UTC().Add(time.Hour)
			if explicit {
				runtime.DaemonID, runtime.DaemonProfile = "daemon-heartbeat-fixture", "desktop-api.riido.ai"
				store.putDaemonLocked(DeviceDaemonRecord{DeviceID: device.DeviceID, DaemonID: runtime.DaemonID, Profile: runtime.DaemonProfile, LastSeenAt: heartbeat})
			} else {
				runtime.DaemonID, runtime.DaemonProfile = "", ""
				device.DaemonLastSeenAt = heartbeat
			}
			return
		}
	}
	t.Fatal("agent runtime not found")
}
