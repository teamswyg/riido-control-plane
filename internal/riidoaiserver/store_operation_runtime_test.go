package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestAssignmentOperationRuntimeCapturesStateMutations(t *testing.T) {
	ctx := context.Background()
	fixedNow := time.Date(2026, 5, 28, 1, 2, 3, 0, time.UTC)
	operations := &runtimeFakeAssignmentOperationStore{}
	store := NewStoreWithConfig(StoreConfig{
		Now:            func() time.Time { return fixedNow },
		OperationStore: operations,
	})
	defer store.Close()

	assignment, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "make hello world",
	})
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	poll, err := store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if poll.Action != PollStart || poll.Assignment == nil {
		t.Fatalf("poll = %+v", poll)
	}
	if _, err := store.RecordAgentEvent(ctx, "agent-1", AgentEventRequest{
		AssignmentID: assignment.ID,
		TaskID:       "task-a",
		DaemonID:     "daemon-1",
		RuntimeID:    "runtime-1",
		State:        AssignmentReady,
		EventType:    EventAssignmentReady,
		Message:      "ready",
	}); err != nil {
		t.Fatalf("RecordAgentEvent: %v", err)
	}

	if len(operations.records) != 3 {
		t.Fatalf("operation count = %d: %+v", len(operations.records), operations.records)
	}
	want := []struct {
		operationType AssignmentOperationType
		eventType     string
		state         AssignmentState
	}{
		{AssignmentOperationAssignTask, EventAssignmentQueued, AssignmentQueued},
		{AssignmentOperationPollStart, EventAssignmentLeased, AssignmentLeased},
		{AssignmentOperationAgentEvent, EventAssignmentReady, AssignmentReady},
	}
	for i, want := range want {
		record := operations.records[i]
		if record.SchemaVersion != AssignmentOperationSchemaVersion || record.OperationType != want.operationType {
			t.Fatalf("record[%d] header = %+v", i, record)
		}
		if record.TaskID != "task-a" || record.AssignmentID != assignment.ID || record.AgentID != "agent-1" {
			t.Fatalf("record[%d] identity = %+v", i, record)
		}
		if len(record.Events) != 1 || record.Events[0].Type != want.eventType || record.Events[0].State != want.state {
			t.Fatalf("record[%d] events = %+v", i, record.Events)
		}
		if record.OperationID == "" || !record.RecordedAt.Equal(fixedNow) {
			t.Fatalf("record[%d] operation metadata = %+v", i, record)
		}
	}
}

func TestPollAgentUsesAssignmentClaimerWithoutDuplicateOperationSave(t *testing.T) {
	ctx := context.Background()
	fixedNow := time.Date(2026, 5, 28, 1, 2, 3, 0, time.UTC)
	operations := &runtimeFakeClaimingAssignmentOperationStore{}
	store := NewStoreWithConfig(StoreConfig{
		Now:            func() time.Time { return fixedNow },
		OperationStore: operations,
	})
	defer store.Close()

	assignment, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "make hello world",
	})
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	claimed := assignment
	claimed.State = AssignmentLeased
	claimed.LeaseToken = "lease-from-durable-store"
	claimed.UpdatedAt = fixedNow
	events := []TaskEvent{{
		Seq:          2,
		TaskID:       claimed.TaskID,
		AssignmentID: claimed.ID,
		AgentID:      claimed.AgentID,
		Type:         EventAssignmentLeased,
		State:        AssignmentLeased,
		Metadata:     map[string]string{"lease_token": claimed.LeaseToken},
		At:           fixedNow,
	}}
	operations.claim = AssignmentClaimResult{
		Claimed:    true,
		Assignment: claimed,
		Operation: AssignmentOperationRecord{
			SchemaVersion: AssignmentOperationSchemaVersion,
			OperationID:   assignmentOperationID(AssignmentOperationPollStart, claimed, events),
			OperationType: AssignmentOperationPollStart,
			TaskID:        claimed.TaskID,
			AssignmentID:  claimed.ID,
			AgentID:       claimed.AgentID,
			Assignment:    claimed,
			Events:        events,
			RecordedAt:    fixedNow,
		},
	}

	poll, err := store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if poll.Action != PollStart || poll.Assignment == nil || poll.Assignment.LeaseToken != claimed.LeaseToken {
		t.Fatalf("poll = %+v", poll)
	}
	if operations.claimCalls != 1 || operations.claimAgentID != "agent-1" {
		t.Fatalf("claim calls = %d agent=%q", operations.claimCalls, operations.claimAgentID)
	}
	if len(operations.records) != 1 {
		t.Fatalf("operation store should only contain assign operation, got %d: %+v", len(operations.records), operations.records)
	}
	history, _, cancel, err := store.SubscribeTask(ctx, "task-a")
	if err != nil {
		t.Fatalf("SubscribeTask: %v", err)
	}
	cancel()
	if len(history) != 2 || history[1].Type != EventAssignmentLeased || history[1].Seq != 2 {
		t.Fatalf("history = %+v", history)
	}
}

func TestOpenStoreWithConfigReplaysAssignmentOperationsWithoutSnapshot(t *testing.T) {
	ctx := context.Background()
	fixedNow := time.Date(2026, 5, 28, 1, 2, 3, 0, time.UTC)
	operations := &runtimeFakeAssignmentOperationStore{}
	source := NewStoreWithConfig(StoreConfig{
		Now:            func() time.Time { return fixedNow },
		OperationStore: operations,
	})
	assignment, err := source.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "make hello world",
	})
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	if _, err := source.PollAgent(ctx, "agent-1", daemonPollRequest()); err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if _, err := source.RecordAgentEvent(ctx, "agent-1", AgentEventRequest{
		AssignmentID: assignment.ID,
		TaskID:       "task-a",
		DaemonID:     "daemon-1",
		RuntimeID:    "runtime-1",
		State:        AssignmentReady,
		EventType:    EventAssignmentReady,
		Message:      "ready",
	}); err != nil {
		t.Fatalf("RecordAgentEvent: %v", err)
	}
	source.Close()

	reloaded, err := OpenStoreWithConfig(ctx, StoreConfig{OperationStore: operations})
	if err != nil {
		t.Fatalf("OpenStoreWithConfig: %v", err)
	}
	defer reloaded.Close()
	history, _, cancel, err := reloaded.SubscribeTask(ctx, "task-a")
	if err != nil {
		t.Fatalf("SubscribeTask: %v", err)
	}
	cancel()
	if len(history) != 3 || history[0].Type != EventAssignmentQueued || history[1].Type != EventAssignmentLeased || history[2].Type != EventAssignmentReady {
		t.Fatalf("history = %+v", history)
	}
	poll, err := reloaded.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent reloaded: %v", err)
	}
	if poll.Action != PollActive || poll.Assignment == nil || poll.Assignment.State != AssignmentReady {
		t.Fatalf("poll reloaded = %+v", poll)
	}
}

func TestOpenStoreWithConfigOverlaysAssignmentProjections(t *testing.T) {
	ctx := context.Background()
	createdAt := time.Date(2026, 6, 18, 1, 2, 3, 0, time.UTC)
	running := Assignment{
		ID:              "asn-000001",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "make hello world",
		State:           AssignmentRunning,
		CreatedAt:       createdAt,
		UpdatedAt:       createdAt.Add(time.Minute),
	}
	completed := running
	completed.State = AssignmentCompleted
	completed.UpdatedAt = createdAt.Add(2 * time.Minute)
	operations := &runtimeFakeProjectionReaderOperationStore{
		runtimeFakeAssignmentOperationStore: runtimeFakeAssignmentOperationStore{records: []AssignmentOperationRecord{{
			SchemaVersion: AssignmentOperationSchemaVersion,
			OperationID:   "agent-event:asn-000001:1",
			OperationType: AssignmentOperationAgentEvent,
			TaskID:        running.TaskID,
			AssignmentID:  running.ID,
			AgentID:       running.AgentID,
			Assignment:    running,
			Events: []TaskEvent{{
				Seq:          1,
				TaskID:       running.TaskID,
				AssignmentID: running.ID,
				AgentID:      running.AgentID,
				Type:         EventAssignmentRunning,
				State:        AssignmentRunning,
				At:           running.UpdatedAt,
			}},
			RecordedAt: running.UpdatedAt,
		}}},
		projections: map[string]AssignmentProjection{
			completed.ID: {
				Assignment:   completed,
				LastEventSeq: 2,
			},
		},
	}
	reloaded, err := OpenStoreWithConfig(ctx, StoreConfig{OperationStore: operations})
	if err != nil {
		t.Fatalf("OpenStoreWithConfig: %v", err)
	}
	defer reloaded.Close()
	metrics, err := reloaded.Metrics(ctx)
	if err != nil {
		t.Fatalf("Metrics: %v", err)
	}
	if metrics.AssignmentsByState[AssignmentRunning] != 0 || metrics.AssignmentsByState[AssignmentCompleted] != 1 {
		t.Fatalf("metrics states = %+v", metrics.AssignmentsByState)
	}
	poll, err := reloaded.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if poll.Action != PollNone || poll.Assignment != nil {
		t.Fatalf("poll after projection overlay = %+v", poll)
	}
}

func TestHeartbeatAgentRefreshesDurableActiveLease(t *testing.T) {
	now := time.Date(2026, 5, 28, 1, 2, 3, 0, time.UTC)
	operations := &runtimeFakeActiveLeaseOperationStore{}
	store := NewStoreWithConfig(StoreConfig{
		Now:                 func() time.Time { return now },
		ActiveLeaseDuration: time.Minute,
		OperationStore:      operations,
	})
	defer store.Close()
	ctx := context.Background()

	assigned, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "make hello world",
	})
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	poll, err := store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent start: %v", err)
	}
	if poll.Action != PollStart || poll.Assignment == nil || poll.Assignment.ID != assigned.ID {
		t.Fatalf("poll start = %+v", poll)
	}

	now = now.Add(30 * time.Second)
	heartbeat, err := store.HeartbeatAgent(ctx, "agent-1", AgentHeartbeatRequest{
		DaemonID:            "daemon-1",
		RuntimeID:           "runtime-1",
		ActiveAssignmentIDs: []string{assigned.ID},
	})
	if err != nil {
		t.Fatalf("HeartbeatAgent: %v", err)
	}
	if len(heartbeat.RefreshedAssignments) != 1 || len(operations.refreshes) != 1 {
		t.Fatalf("heartbeat=%+v refreshes=%+v", heartbeat, operations.refreshes)
	}

	now = now.Add(40 * time.Second)
	poll, err = store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent active: %v", err)
	}
	if poll.Action != PollActive || poll.Assignment == nil || poll.Assignment.ID != assigned.ID {
		t.Fatalf("poll after heartbeat = %+v", poll)
	}
	if operations.loadCalls == 0 {
		t.Fatal("active poll should check durable active lease for cross-process cancellation")
	}
}

func TestHeartbeatAgentFailsStaleActiveLeaseBeforeRefresh(t *testing.T) {
	now := time.Date(2026, 5, 28, 1, 2, 3, 0, time.UTC)
	operations := &runtimeFakeActiveLeaseOperationStore{}
	store := NewStoreWithConfig(StoreConfig{
		Now:            func() time.Time { return now },
		OperationStore: operations,
	})
	defer store.Close()
	ctx := context.Background()

	first, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "first",
	})
	if err != nil {
		t.Fatalf("AssignTask first: %v", err)
	}
	if _, err := store.AssignTask(ctx, "task-b", AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "second",
	}); err != nil {
		t.Fatalf("AssignTask second: %v", err)
	}
	poll, err := store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent start: %v", err)
	}
	if poll.Action != PollStart || poll.Assignment == nil || poll.Assignment.ID != first.ID {
		t.Fatalf("poll start = %+v", poll)
	}

	now = now.Add(time.Duration(DefaultAssignmentActiveLeaseSeconds)*time.Second + time.Second)
	heartbeat, err := store.HeartbeatAgent(ctx, "agent-1", AgentHeartbeatRequest{
		DaemonID:            "daemon-1",
		RuntimeID:           "runtime-1",
		ActiveAssignmentIDs: []string{first.ID},
	})
	if err != nil {
		t.Fatalf("HeartbeatAgent stale: %v", err)
	}
	if len(heartbeat.RefreshedAssignments) != 0 || len(operations.refreshes) != 0 {
		t.Fatalf("stale heartbeat must not refresh active lease: heartbeat=%+v refreshes=%+v", heartbeat, operations.refreshes)
	}
	last := operations.records[len(operations.records)-1]
	if last.OperationType != AssignmentOperationAgentEvent || last.Assignment.ID != first.ID || last.Assignment.State != AssignmentFailed {
		t.Fatalf("stale heartbeat operation = %+v", last)
	}

	poll, err = store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent second: %v", err)
	}
	if poll.Action != PollStart || poll.Assignment == nil || poll.Assignment.TaskID != "task-b" {
		t.Fatalf("poll second = %+v", poll)
	}
}

func TestPollAgentFailsStaleActiveAssignmentBeforeNextClaim(t *testing.T) {
	now := time.Date(2026, 5, 28, 1, 2, 3, 0, time.UTC)
	operations := &runtimeFakeActiveLeaseOperationStore{}
	store := NewStoreWithConfig(StoreConfig{
		Now:                 func() time.Time { return now },
		ActiveLeaseDuration: time.Minute,
		OperationStore:      operations,
	})
	defer store.Close()
	ctx := context.Background()

	first, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "first",
	})
	if err != nil {
		t.Fatalf("AssignTask first: %v", err)
	}
	if _, err := store.AssignTask(ctx, "task-b", AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "second",
	}); err != nil {
		t.Fatalf("AssignTask second: %v", err)
	}
	poll, err := store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent start: %v", err)
	}
	if poll.Action != PollStart || poll.Assignment == nil || poll.Assignment.ID != first.ID {
		t.Fatalf("poll start = %+v", poll)
	}
	operations.activeFound = true
	operations.activeLease = AssignmentActiveLease{
		AgentID:            "agent-1",
		ActiveAssignmentID: first.ID,
		LeaseToken:         poll.Assignment.LeaseToken,
		HeartbeatAt:        now,
		LeaseExpiresAt:     now.Add(time.Minute),
		LeaseExpiresUnixMS: now.Add(time.Minute).UnixMilli(),
	}

	now = now.Add(2 * time.Minute)
	poll, err = store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent stale: %v", err)
	}
	if poll.Action != PollNone {
		t.Fatalf("stale poll should only fail old assignment, got %+v", poll)
	}
	last := operations.records[len(operations.records)-1]
	if last.OperationType != AssignmentOperationAgentEvent || last.Assignment.ID != first.ID || last.Assignment.State != AssignmentFailed {
		t.Fatalf("stale operation = %+v", last)
	}

	poll, err = store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent second: %v", err)
	}
	if poll.Action != PollStart || poll.Assignment == nil || poll.Assignment.TaskID != "task-b" {
		t.Fatalf("poll second = %+v", poll)
	}
}

func TestPollAgentUsesDurableCancellingProjection(t *testing.T) {
	now := time.Date(2026, 5, 28, 1, 2, 3, 0, time.UTC)
	assignment := Assignment{
		ID:              "asn-000001",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "first",
		State:           AssignmentCancelling,
		LeaseToken:      "lease-1",
		CreatedAt:       now.Add(-time.Minute),
		UpdatedAt:       now,
	}
	operations := &runtimeFakeActiveLeaseOperationStore{
		activeFound: true,
		activeLease: AssignmentActiveLease{
			AgentID:            assignment.AgentID,
			ActiveAssignmentID: assignment.ID,
			LeaseToken:         assignment.LeaseToken,
			HeartbeatAt:        now,
			LeaseExpiresAt:     now.Add(time.Minute),
			LeaseExpiresUnixMS: now.Add(time.Minute).UnixMilli(),
		},
		projections: map[string]AssignmentProjection{
			assignment.ID: {Assignment: assignment, LastEventSeq: 4},
		},
	}
	store := NewStoreWithConfig(StoreConfig{
		Now:                 func() time.Time { return now },
		ActiveLeaseDuration: time.Minute,
		OperationStore:      operations,
	})
	defer store.Close()

	poll, err := store.PollAgent(context.Background(), "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if poll.Action != PollCancel || poll.Assignment == nil || poll.Assignment.ID != assignment.ID {
		t.Fatalf("durable cancellation poll = %+v", poll)
	}
	if operations.projectionLoadCalls != 1 {
		t.Fatalf("projectionLoadCalls=%d", operations.projectionLoadCalls)
	}
}

type runtimeFakeAssignmentOperationStore struct {
	records []AssignmentOperationRecord
	closed  bool
}

func (s *runtimeFakeAssignmentOperationStore) SaveAssignmentOperation(_ context.Context, record AssignmentOperationRecord) error {
	s.records = append(s.records, record)
	return nil
}

func (s *runtimeFakeAssignmentOperationStore) LoadAssignmentOperations(context.Context) ([]AssignmentOperationRecord, error) {
	return append([]AssignmentOperationRecord(nil), s.records...), nil
}

func (s *runtimeFakeAssignmentOperationStore) Close() error {
	s.closed = true
	return nil
}

type runtimeFakeProjectionReaderOperationStore struct {
	runtimeFakeAssignmentOperationStore
	projections map[string]AssignmentProjection
}

func (s *runtimeFakeProjectionReaderOperationStore) LoadAssignmentProjection(_ context.Context, assignmentID string) (AssignmentProjection, bool, error) {
	projection, ok := s.projections[assignmentID]
	return projection, ok, nil
}

type runtimeFakeClaimingAssignmentOperationStore struct {
	runtimeFakeAssignmentOperationStore
	claim        AssignmentClaimResult
	claimCalls   int
	claimAgentID string
	claimAt      time.Time
}

func (s *runtimeFakeClaimingAssignmentOperationStore) ClaimNextAssignment(_ context.Context, agentID string, at time.Time) (AssignmentClaimResult, error) {
	s.claimCalls++
	s.claimAgentID = agentID
	s.claimAt = at
	return s.claim, nil
}

type runtimeFakeActiveLeaseOperationStore struct {
	runtimeFakeAssignmentOperationStore
	activeLease         AssignmentActiveLease
	activeFound         bool
	loadCalls           int
	refreshes           []Assignment
	projections         map[string]AssignmentProjection
	projectionLoadCalls int
}

func (s *runtimeFakeActiveLeaseOperationStore) LoadAgentActiveAssignment(_ context.Context, agentID string) (AssignmentActiveLease, bool, error) {
	s.loadCalls++
	if !s.activeFound || s.activeLease.AgentID != agentID {
		return AssignmentActiveLease{}, false, nil
	}
	return s.activeLease, true, nil
}

func (s *runtimeFakeActiveLeaseOperationStore) LoadAssignmentProjection(_ context.Context, assignmentID string) (AssignmentProjection, bool, error) {
	s.projectionLoadCalls++
	projection, ok := s.projections[assignmentID]
	return projection, ok, nil
}

func (s *runtimeFakeActiveLeaseOperationStore) RefreshAgentActiveAssignment(_ context.Context, assignment Assignment, at time.Time) error {
	s.refreshes = append(s.refreshes, assignment)
	s.activeFound = true
	s.activeLease = AssignmentActiveLease{
		AgentID:            assignment.AgentID,
		ActiveAssignmentID: assignment.ID,
		LeaseToken:         assignment.LeaseToken,
		HeartbeatAt:        at,
		LeaseExpiresAt:     at.Add(time.Minute),
		LeaseExpiresUnixMS: at.Add(time.Minute).UnixMilli(),
	}
	return nil
}
