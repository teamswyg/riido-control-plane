package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

// memorySplitAIAgentClientSnapshotStore implements both the legacy combined
// interface and the per-collection split interface, tracking per-item save
// counts so tests can assert the daemon hot path never writes the threads item.
type memorySplitAIAgentClientSnapshotStore struct {
	combined     AIAgentClientSnapshot
	combinedOK   bool
	core         AIAgentClientCoreSnapshot
	coreOK       bool
	events       AIAgentClientEventsSnapshot
	eventsOK     bool
	threads      AIAgentClientThreadsSnapshot
	threadsOK    bool
	coreSaves    int
	eventsSaves  int
	threadsSaves int
}

func (s *memorySplitAIAgentClientSnapshotStore) LoadAIAgentClientSnapshot(context.Context) (AIAgentClientSnapshot, bool, error) {
	return s.combined, s.combinedOK, nil
}

func (s *memorySplitAIAgentClientSnapshotStore) SaveAIAgentClientSnapshot(_ context.Context, snapshot AIAgentClientSnapshot) error {
	s.combined = snapshot
	s.combinedOK = true
	return nil
}

func (s *memorySplitAIAgentClientSnapshotStore) LoadCore(context.Context) (AIAgentClientCoreSnapshot, bool, error) {
	return s.core, s.coreOK, nil
}

func (s *memorySplitAIAgentClientSnapshotStore) SaveCore(_ context.Context, core AIAgentClientCoreSnapshot) error {
	s.core = core
	s.coreOK = true
	s.coreSaves++
	return nil
}

func (s *memorySplitAIAgentClientSnapshotStore) LoadEvents(context.Context) (AIAgentClientEventsSnapshot, bool, error) {
	return s.events, s.eventsOK, nil
}

func (s *memorySplitAIAgentClientSnapshotStore) SaveEvents(_ context.Context, events AIAgentClientEventsSnapshot) error {
	s.events = events
	s.eventsOK = true
	s.eventsSaves++
	return nil
}

func (s *memorySplitAIAgentClientSnapshotStore) LoadThreads(context.Context) (AIAgentClientThreadsSnapshot, bool, error) {
	return s.threads, s.threadsOK, nil
}

func (s *memorySplitAIAgentClientSnapshotStore) SaveThreads(_ context.Context, threads AIAgentClientThreadsSnapshot) error {
	s.threads = threads
	s.threadsOK = true
	s.threadsSaves++
	return nil
}

func (s *memorySplitAIAgentClientSnapshotStore) WriteSplitAtomic(_ context.Context, core AIAgentClientCoreSnapshot, events AIAgentClientEventsSnapshot, threads AIAgentClientThreadsSnapshot) error {
	s.core, s.coreOK, s.coreSaves = core, true, s.coreSaves+1
	s.events, s.eventsOK, s.eventsSaves = events, true, s.eventsSaves+1
	s.threads, s.threadsOK, s.threadsSaves = threads, true, s.threadsSaves+1
	return nil
}

func (s *memorySplitAIAgentClientSnapshotStore) WriteCoreEvents(_ context.Context, core AIAgentClientCoreSnapshot, events AIAgentClientEventsSnapshot) error {
	s.core, s.coreOK, s.coreSaves = core, true, s.coreSaves+1
	s.events, s.eventsOK, s.eventsSaves = events, true, s.eventsSaves+1
	return nil
}

func (s *memorySplitAIAgentClientSnapshotStore) resetCounts() {
	s.coreSaves, s.eventsSaves, s.threadsSaves = 0, 0, 0
}

func TestAIAgentClientSplitStoreMigratesLegacyAndRoundTrips(t *testing.T) {
	ctx := context.Background()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-dev"}

	// Build a legacy combined snapshot via the combined store.
	legacy := &memoryAIAgentClientSnapshotStore{}
	seed, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), legacy)
	if err != nil {
		t.Fatalf("open seed store: %v", err)
	}
	created, err := seed.CreateAIAgent(ctx, principal, CreateAgentConfigurationRequest{
		Name:       "Split Agent",
		Visibility: AgentVisibilityPrivate,
		RuntimeID:  "runtime-cursor-dev",
	})
	if err != nil {
		t.Fatalf("CreateAIAgent: %v", err)
	}
	if _, err := seed.SubmitAIAgentTaskComment(ctx, principal, "task-split", SubmitAIAgentTaskCommentRequest{
		AgentID: created.Agent.AgentID,
		Body:    "seed thread",
	}); err != nil {
		t.Fatalf("SubmitAIAgentTaskComment: %v", err)
	}
	if !legacy.ok {
		t.Fatal("legacy combined snapshot was not written")
	}

	// Open a split store seeded with that legacy combined item → migrates.
	split := &memorySplitAIAgentClientSnapshotStore{combined: legacy.snapshot, combinedOK: true}
	if _, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), split); err != nil {
		t.Fatalf("open split store (migrate): %v", err)
	}
	if !split.coreOK || !split.eventsOK || !split.threadsOK {
		t.Fatalf("migration must write all split items: core=%v events=%v threads=%v", split.coreOK, split.eventsOK, split.threadsOK)
	}
	if !split.combinedOK {
		t.Fatal("legacy combined item must be preserved for rollback")
	}

	// Drop the legacy item, then reopen: must load purely from the split items.
	split.combinedOK = false
	reopened, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), split)
	if err != nil {
		t.Fatalf("reopen split store: %v", err)
	}
	bootstrap, err := reopened.BootstrapAIAgentClient(ctx, principal, ClientKindWeb)
	if err != nil {
		t.Fatalf("BootstrapAIAgentClient: %v", err)
	}
	if !agentListContains(bootstrap.Agents, created.Agent.AgentID) {
		t.Fatalf("migrated split state missing created agent: %+v", bootstrap.Agents)
	}
}

func TestAIAgentClientSplitSyncWritesCoreOnlyOnHeartbeat(t *testing.T) {
	ctx := context.Background()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-dev"}
	split := &memorySplitAIAgentClientSnapshotStore{}
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), split)
	if err != nil {
		t.Fatalf("open split store: %v", err)
	}
	enrollment, err := store.EnrollDeviceCredential(ctx, principal, "workspace-dev", EnrollDeviceRequest{DisplayName: "Split Mac"})
	if err != nil {
		t.Fatalf("EnrollDeviceCredential: %v", err)
	}
	req := DeviceRuntimeSnapshotSyncRequest{
		DaemonID:          "daemon-a",
		DeviceID:          enrollment.DeviceID,
		DeviceDisplayName: "Split Mac",
		Profile:           "development",
		PID:               1234,
		UptimeSeconds:     10,
		StartedAt:         time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		Runtimes: []RuntimeSnapshotRecord{{
			RuntimeID: "daemon-a:codex",
			Kind:      RuntimeKindCodex,
			Models:    []RuntimeModelRecord{{ModelID: "gpt-5.5", Label: "GPT-5.5", IsDefault: true}},
		}},
	}

	// First sync: a new device/runtime is a meaningful change → appends an event
	// → writes core+events. Threads must NOT be written by the daemon hot path.
	split.resetCounts()
	if _, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, req); err != nil {
		t.Fatalf("first Sync: %v", err)
	}
	if split.threadsSaves != 0 {
		t.Fatalf("Sync must never write the threads item, got %d", split.threadsSaves)
	}
	if split.coreSaves == 0 {
		t.Fatal("Sync must write the core item")
	}
	if split.eventsSaves == 0 {
		t.Fatal("first Sync (new device) appends an event → must write the events item")
	}

	// Second identical sync: a plain heartbeat (no meaningful change) → core only.
	split.resetCounts()
	if _, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, req); err != nil {
		t.Fatalf("heartbeat Sync: %v", err)
	}
	if split.coreSaves == 0 {
		t.Fatal("heartbeat must still write core (daemon last-seen)")
	}
	if split.eventsSaves != 0 {
		t.Fatalf("heartbeat (no change) must not write the events item, got %d", split.eventsSaves)
	}
	if split.threadsSaves != 0 {
		t.Fatalf("heartbeat must never write the threads item, got %d", split.threadsSaves)
	}
}
