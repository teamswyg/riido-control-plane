package riidoaiserver

import "testing"

func TestPruneExpectedDaemonProfilesForSnapshotLocked(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	deviceID := "device-prune-unit"
	store.mu.Lock()
	store.devices = append(store.devices, DeviceRecord{
		DeviceID: deviceID,
		Runtimes: []RuntimeRecord{
			profileRuntime(deviceID, "staging", RuntimeKindCodex),
			profileRuntime(deviceID, "production", RuntimeKindClaudeCode),
		},
	})
	store.daemons["daemon-staging"] = profileDaemon(deviceID, "staging")
	store.daemons["daemon-production"] = profileDaemon(deviceID, "production")
	pruned := store.pruneExpectedDaemonProfilesForSnapshotLocked(deviceID, "production")
	store.mu.Unlock()

	if pruned.daemons != 1 || pruned.runtimes != 1 || !pruned.changed() {
		t.Fatalf("pruned = %+v", pruned)
	}
	assertOnlyProfileRuntime(t, store, deviceID, "production")
	if got := countDaemonsForDevice(store, deviceID); got != 1 {
		t.Fatalf("daemon records = %d, want 1", got)
	}
}

func TestPruneExpectedDaemonProfilesForSnapshotNoExpectedProfile(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	store.mu.Lock()
	pruned := store.pruneExpectedDaemonProfilesForSnapshotLocked("device-any", "")
	store.mu.Unlock()
	if pruned.changed() {
		t.Fatalf("empty expected profile should not prune: %+v", pruned)
	}
}

func TestPruneUnexpectedDaemonProfilesNoopBranches(t *testing.T) {
	if (*DevelopmentAIAgentClientStore)(nil).pruneUnexpectedDaemonProfiles("unit") {
		t.Fatal("nil store should not prune")
	}
	store := NewDevelopmentAIAgentClientStore()
	if store.pruneUnexpectedDaemonProfiles("unit") {
		t.Fatal("empty expected profile should not prune")
	}
	store.mu.Lock()
	pruned := store.pruneUnexpectedDaemonProfilesLocked("device-any", "")
	store.mu.Unlock()
	if pruned.changed() {
		t.Fatalf("empty expected profile should not prune locked path: %+v", pruned)
	}
}
