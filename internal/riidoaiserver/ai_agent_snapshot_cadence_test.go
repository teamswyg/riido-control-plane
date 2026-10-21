package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestSnapshotCadenceCannotOutliveRuntimeLiveness(t *testing.T) {
	store, err := OpenPersistentAIAgentClientStore(context.Background(), nil, &memoryAIAgentClientSnapshotStore{})
	if err != nil {
		t.Fatal(err)
	}
	for _, intervals := range [][2]time.Duration{
		{time.Minute, 5 * time.Minute},
		{30 * time.Second, 15 * time.Second},
		{0, 30 * time.Second},
		{30 * time.Second, 0},
	} {
		if err := store.ConfigureSnapshotCadence(intervals[0], intervals[1]); err == nil {
			t.Errorf("accepted cadence %v that can make a live runtime stale", intervals)
		}
		if store.snapshotReloadInterval != 15*time.Second || store.snapshotHeartbeatSaveInterval != 15*time.Second {
			t.Fatal("invalid cadence partially applied")
		}
	}
}

func TestSnapshotLocalWritesDoNotDelayOtherServerRefresh(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store, err := OpenPersistentAIAgentClientStore(ctx, nil, snapshots)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	store.snapshotClock = func() time.Time { return now }
	store.snapshotLastReloadAt = now
	principal := profileColorTestPrincipal()
	now = now.Add(10 * time.Second)
	if _, err := store.UpdateAIAgentConfiguration(ctx, principal, "agent-owned-claude", UpdateAgentConfigurationRequest{Name: "Local edit"}); err != nil {
		t.Fatal(err)
	}
	loads := snapshots.loads
	now = now.Add(6 * time.Second)
	if _, err := store.BootstrapAIAgentClient(ctx, principal, ClientKindWeb); err != nil {
		t.Fatal(err)
	}
	if snapshots.loads != loads+1 {
		t.Fatal("local saves postponed reading other servers' runtime heartbeats")
	}
}

func TestRuntimeHeartbeatFailedSaveRemainsDue(t *testing.T) {
	ctx := context.Background()
	snapshots := &failingPresenceCommitStore{grantsReadingSnapshotStore: grantsReadingSnapshotStore{&memoryAIAgentClientSnapshotStore{}}}
	store, err := OpenPersistentAIAgentClientStore(ctx, nil, snapshots)
	if err != nil {
		t.Fatal(err)
	}
	req := DeviceRuntimeSnapshotSyncRequest{DeviceID: "retry-device", DaemonID: "retry-daemon",
		Runtimes: []RuntimeSnapshotRecord{{RuntimeID: "retry-runtime", Kind: RuntimeKindCodex}}}
	if _, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, profileColorTestPrincipal(), req); err != nil {
		t.Fatal(err)
	}
	key := daemonStorageKeyFor(req.DeviceID, "", req.DaemonID)
	lastCommitted := time.Now().Add(-time.Minute)
	store.snapshotDaemonLastSeen[key] = lastCommitted
	snapshots.fail = true
	if _, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, profileColorTestPrincipal(), req); err == nil {
		t.Fatal("heartbeat save failure was hidden")
	}
	if !store.snapshotDaemonLastSeen[key].Equal(lastCommitted) {
		t.Fatal("failed save advanced the committed heartbeat")
	}
	snapshots.fail = false
	saves := snapshots.saves
	if _, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, profileColorTestPrincipal(), req); err != nil {
		t.Fatal(err)
	}
	if snapshots.saves != saves+1 || !store.snapshotDaemonLastSeen[key].After(lastCommitted) {
		t.Fatal("unchanged heartbeat did not retry its failed save")
	}
}
