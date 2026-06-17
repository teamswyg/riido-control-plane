package riidoaiserver

import "time"

func prepareRuntimeBindingCandidateFixture(store *DevelopmentAIAgentClientStore, now time.Time) {
	store.mu.Lock()
	defer store.mu.Unlock()
	makeDefaultDeviceFresh(store, now)
	store.agents["agent-cursor-idle"] = cursorAgentFixture("agent-cursor-idle", now.Add(-time.Minute), AgentWorkStatusIdle, 0)
	store.agents["agent-cursor-active"] = cursorAgentFixture("agent-cursor-active", now, AgentWorkStatusRunning, 1)
	store.taskThreads["task-cursor-active"] = []AIAgentTaskThreadRecord{cursorActiveThreadFixture(now)}
}

func makeDefaultDeviceFresh(store *DevelopmentAIAgentClientStore, now time.Time) {
	for i := range store.devices {
		if store.devices[i].DeviceID != "device-dev-macbook" {
			continue
		}
		store.devices[i].DaemonLastSeenAt = now
		for j := range store.devices[i].Runtimes {
			store.devices[i].Runtimes[j].LastDetectedAt = now
			store.devices[i].Runtimes[j].Availability = RuntimeAvailabilityOnline
			store.devices[i].Runtimes[j].DetectionState = RuntimeDetectionStateDetected
		}
	}
	daemon := store.daemons["device-dev-macbook"]
	daemon.LastSeenAt = now
	daemon.Availability = DaemonAvailabilityOnline
	store.daemons["device-dev-macbook"] = daemon
}
