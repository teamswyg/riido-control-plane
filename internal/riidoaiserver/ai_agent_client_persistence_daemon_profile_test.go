package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestPersistentAIAgentClientStorePrunesUnexpectedDaemonProfilesOnRestore(t *testing.T) {
	ctx := context.Background()
	seed := NewDevelopmentAIAgentClientStore()
	if err := seed.ConfigureDaemonProfile("production"); err != nil {
		t.Fatalf("ConfigureDaemonProfile: %v", err)
	}
	snapshots := &memoryAIAgentClientSnapshotStore{
		ok:       true,
		snapshot: daemonProfileRestoreSnapshot(),
	}
	store, err := OpenPersistentAIAgentClientStore(ctx, seed, snapshots)
	if err != nil {
		t.Fatalf("OpenPersistentAIAgentClientStore: %v", err)
	}
	assertOnlyProfileRuntime(t, store.DevelopmentAIAgentClientStore,
		"device-profile-restore", "production")
	if got := countDaemonsForDevice(store.DevelopmentAIAgentClientStore,
		"device-profile-restore"); got != 1 {
		t.Fatalf("daemon records = %d, want 1", got)
	}
	if snapshots.saves == 0 {
		t.Fatal("restore prune must save the cleaned snapshot")
	}
}

func daemonProfileRestoreSnapshot() AIAgentClientSnapshot {
	return AIAgentClientSnapshot{
		SchemaVersion: AIAgentClientPersistenceSchemaVersion,
		SavedAt:       time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC),
		Devices:       []DeviceRecord{daemonProfileRestoreDevice()},
		Daemons: []DeviceDaemonRecord{
			profileDaemon("device-profile-restore", "staging"),
			profileDaemon("device-profile-restore", "production"),
		},
		Agents:             []AgentClientRecord{},
		Fixtures:           []AgentOnboardingFixture{},
		TaskThreads:        map[string][]AIAgentTaskThreadRecord{},
		TaskThreadMessages: map[string][]AIAgentTaskThreadHistoryMessage{},
		Events:             []AIAgentClientEventSnapshot{},
		NextDaemonCommand:  1,
	}
}

func daemonProfileRestoreDevice() DeviceRecord {
	return DeviceRecord{
		DeviceID:         "device-profile-restore",
		OwnerPrincipalID: "user-profiled",
		Runtimes: []RuntimeRecord{
			profileRuntime("device-profile-restore", "staging", RuntimeKindCodex),
			profileRuntime("device-profile-restore", "production", RuntimeKindClaudeCode),
		},
	}
}
