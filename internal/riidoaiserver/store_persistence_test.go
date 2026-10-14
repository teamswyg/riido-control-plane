package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFileOutboxPersistsTaskEvents(t *testing.T) {
	outboxPath := filepath.Join(t.TempDir(), "riido-ai-server", "outbox.jsonl")
	outbox, err := NewFileOutbox(outboxPath)
	if err != nil {
		t.Fatalf("NewFileOutbox: %v", err)
	}
	store := NewStoreWithConfig(StoreConfig{Outbox: outbox})
	ctx := context.Background()

	if _, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "make hello world",
	}); err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	if _, err := store.PollAgent(ctx, "agent-1", daemonPollRequest()); err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	store.Close()

	records := readOutboxRecords(t, outboxPath)
	if len(records) != 2 {
		t.Fatalf("outbox record count = %d: %+v", len(records), records)
	}
	if records[0].SchemaVersion != OutboxRecordSchemaVersion || records[0].Event.Type != EventAssignmentQueued {
		t.Fatalf("first outbox record = %+v", records[0])
	}
	if records[1].SchemaVersion != OutboxRecordSchemaVersion || records[1].Event.Type != EventAssignmentLeased {
		t.Fatalf("second outbox record = %+v", records[1])
	}
}

func TestOutboxFailureRecordsMetricsWithoutFailingMutation(t *testing.T) {
	ctx := context.Background()
	outbox := &failingEventSink{err: errors.New("outbox unavailable")}
	store := NewStoreWithConfig(StoreConfig{Outbox: outbox})
	defer store.Close()

	if _, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "make hello world",
	}); err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	metrics, err := store.Metrics(ctx)
	if err != nil {
		t.Fatalf("Metrics: %v", err)
	}
	if metrics.OutboxErrorsTotal != 1 || metrics.EventAppendLatencySamplesTotal != 1 {
		t.Fatalf("outbox metrics = %+v", metrics)
	}
}

func TestFileStoreSnapshotRestoresAssignmentsAndEvents(t *testing.T) {
	ctx := context.Background()
	snapshotPath := filepath.Join(t.TempDir(), "riido-ai-server", "store-snapshot.json")
	snapshot, err := NewFileStoreSnapshot(snapshotPath)
	if err != nil {
		t.Fatalf("NewFileStoreSnapshot: %v", err)
	}
	store, err := OpenStoreWithConfig(ctx, StoreConfig{SnapshotStore: snapshot})
	if err != nil {
		t.Fatalf("OpenStoreWithConfig: %v", err)
	}

	assignment, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "make hello world",
	})
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	if _, err := store.PollAgent(ctx, "agent-1", daemonPollRequest()); err != nil {
		t.Fatalf("PollAgent: %v", err)
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
	store.Close()

	reloadedSnapshot, err := NewFileStoreSnapshot(snapshotPath)
	if err != nil {
		t.Fatalf("NewFileStoreSnapshot reload: %v", err)
	}
	reloaded, err := OpenStoreWithConfig(ctx, StoreConfig{SnapshotStore: reloadedSnapshot})
	if err != nil {
		t.Fatalf("OpenStoreWithConfig reload: %v", err)
	}
	defer reloaded.Close()

	history, _, cancel, err := reloaded.SubscribeTask(ctx, "task-a")
	if err != nil {
		t.Fatalf("SubscribeTask reloaded: %v", err)
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
	metrics, err := reloaded.Metrics(ctx)
	if err != nil {
		t.Fatalf("Metrics: %v", err)
	}
	if metrics.AssignmentsByState[AssignmentReady] != 1 || metrics.TaskEventsTotal != 3 {
		t.Fatalf("metrics = %+v", metrics)
	}
	if metrics.PollActionsTotal[PollStart] != 1 || metrics.AgentEventsTotal != 1 || metrics.EventAppendLatencySamplesTotal != 3 {
		t.Fatalf("reloaded metrics did not preserve counters: %+v", metrics)
	}
}

func TestStoreSnapshotRebuildsMissingMetricsFromEventHistory(t *testing.T) {
	at := time.Date(2026, 6, 16, 7, 0, 0, 0, time.UTC)
	snapshot := StoreSnapshot{
		SchemaVersion: StoreSnapshotSchemaVersion,
		SavedAt:       at,
		Tasks: []StoreSnapshotTask{{
			ID:                  "task-a",
			ComponentID:         "component-1",
			CurrentAssignmentID: "asn-000001",
		}},
		Assignments: []Assignment{{
			ID:              "asn-000001",
			TaskID:          "task-a",
			ComponentID:     "component-1",
			AgentID:         "agent-1",
			RuntimeProvider: "codex",
			Prompt:          "make hello world",
			State:           AssignmentRunning,
			CreatedAt:       at,
			UpdatedAt:       at,
		}},
		AgentAssignments: map[string][]string{"agent-1": {"asn-000001"}},
		Events: map[string][]TaskEvent{"task-a": {
			{Seq: 1, TaskID: "task-a", AssignmentID: "asn-000001", AgentID: "agent-1", Type: EventAssignmentQueued, State: AssignmentQueued, At: at},
			{Seq: 2, TaskID: "task-a", AssignmentID: "asn-000001", AgentID: "agent-1", Type: EventAssignmentLeased, State: AssignmentLeased, At: at},
			{Seq: 3, TaskID: "task-a", AssignmentID: "asn-000001", AgentID: "agent-1", Type: EventAssignmentRunning, State: AssignmentRunning, At: at},
		}},
		NextAssignmentSeq: 1,
		NextEventSeq:      3,
	}

	state, err := stateFromSnapshot(snapshot)
	if err != nil {
		t.Fatalf("stateFromSnapshot: %v", err)
	}
	if state.pollActionsTotal[PollStart] != 1 || state.pollRequestsTotal != 1 {
		t.Fatalf("poll metrics = requests:%d actions:%+v", state.pollRequestsTotal, state.pollActionsTotal)
	}
	if state.agentEventsTotal != 1 ||
		state.eventAppendLatency.samplesTotal != 3 ||
		state.eventAppendLatency.totalMilliseconds != 3 ||
		state.eventAppendLatency.maxMilliseconds != 1 ||
		state.eventAppendLatency.lastMilliseconds != 1 {
		t.Fatalf("event metrics = agent:%d latency:%+v", state.agentEventsTotal, state.eventAppendLatency)
	}
}

func TestDecodeStoreSnapshotRejectsUnknownFieldsAndTrailingData(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "unknown field",
			raw:  `{"schema_version":"riido-ai-server-store-snapshot.v1","saved_at":"2026-05-28T00:00:00Z","tasks":[],"assignments":[],"agent_assignments":{},"events":{},"next_assignment_seq":0,"next_event_seq":0,"extra":true}`,
			want: "unknown field",
		},
		{
			name: "trailing data",
			raw:  `{"schema_version":"riido-ai-server-store-snapshot.v1","saved_at":"2026-05-28T00:00:00Z","tasks":[],"assignments":[],"agent_assignments":{},"events":{},"next_assignment_seq":0,"next_event_seq":0} {}`,
			want: "trailing data",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := decodeStoreSnapshot(strings.NewReader(tc.raw)); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("decodeStoreSnapshot err=%v, want %q", err, tc.want)
			}
		})
	}
}

func TestStateFromSnapshotRejectsInvalidSnapshotState(t *testing.T) {
	valid := StoreSnapshot{
		SchemaVersion:     StoreSnapshotSchemaVersion,
		Tasks:             []StoreSnapshotTask{{ID: "task-a"}},
		Assignments:       []Assignment{{ID: "asn-000001", TaskID: "task-a", AgentID: "agent-1", State: AssignmentQueued}},
		AgentAssignments:  map[string][]string{"agent-1": {"asn-000001"}},
		Events:            map[string][]TaskEvent{},
		NextAssignmentSeq: 1,
		NextEventSeq:      0,
	}

	for _, tc := range []struct {
		name string
		edit func(*StoreSnapshot)
		want string
	}{
		{
			name: "unsupported schema",
			edit: func(snapshot *StoreSnapshot) { snapshot.SchemaVersion = "riido-ai-server-store-snapshot.v0" },
			want: "unsupported store snapshot schema_version",
		},
		{
			name: "blank task id",
			edit: func(snapshot *StoreSnapshot) { snapshot.Tasks[0].ID = " " },
			want: "task id is required",
		},
		{
			name: "blank assignment id",
			edit: func(snapshot *StoreSnapshot) { snapshot.Assignments[0].ID = "" },
			want: "assignment id is required",
		},
		{
			name: "unknown assignment reference",
			edit: func(snapshot *StoreSnapshot) { snapshot.AgentAssignments["agent-1"] = []string{"asn-missing"} },
			want: "references unknown assignment",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := valid
			snapshot.Tasks = append([]StoreSnapshotTask(nil), valid.Tasks...)
			snapshot.Assignments = append([]Assignment(nil), valid.Assignments...)
			snapshot.AgentAssignments = map[string][]string{"agent-1": append([]string(nil), valid.AgentAssignments["agent-1"]...)}
			snapshot.Events = map[string][]TaskEvent{}
			tc.edit(&snapshot)
			if _, err := stateFromSnapshot(snapshot); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("stateFromSnapshot err=%v, want %q", err, tc.want)
			}
		})
	}
}

type failingEventSink struct {
	err    error
	events []TaskEvent
	closed bool
}

func (s *failingEventSink) AppendTaskEvent(_ context.Context, event TaskEvent) error {
	s.events = append(s.events, event)
	return s.err
}

func (s *failingEventSink) Close() error {
	s.closed = true
	return nil
}

func readOutboxRecords(t *testing.T, path string) []OutboxRecord {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open outbox: %v", err)
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	var records []OutboxRecord
	for {
		var record OutboxRecord
		err := dec.Decode(&record)
		if errors.Is(err, io.EOF) {
			return records
		}
		if err != nil {
			t.Fatalf("decode outbox: %v", err)
		}
		records = append(records, record)
	}
}

func TestFileStoreSnapshotRejectsInvalidConfig(t *testing.T) {
	if _, err := NewFileStoreSnapshot(" "); err == nil {
		t.Fatal("expected missing snapshot path error")
	}
	if _, err := NewFileOutbox(" "); err == nil {
		t.Fatal("expected missing outbox path error")
	}
}
