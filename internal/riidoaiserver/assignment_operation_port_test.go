package riidoaiserver

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestAssignmentOperationRecordValidation(t *testing.T) {
	now := time.Date(2026, 5, 27, 13, 0, 0, 0, time.UTC)
	base := validAssignmentOperationRecord(now)
	cases := []struct {
		name string
		edit func(*AssignmentOperationRecord)
		want string
	}{
		{name: "valid"},
		{name: "unsupported schema", edit: func(record *AssignmentOperationRecord) {
			record.SchemaVersion = "riido-ai-server-assignment-operation.v0"
		}, want: "unsupported assignment operation schema_version"},
		{name: "missing operation id", edit: func(record *AssignmentOperationRecord) { record.OperationID = " " }, want: "assignment operation_id is required"},
		{name: "missing operation type", edit: func(record *AssignmentOperationRecord) { record.OperationType = "" }, want: "assignment operation_type is required"},
		{name: "missing task id", edit: func(record *AssignmentOperationRecord) { record.TaskID = " " }, want: "assignment operation task_id is required"},
		{name: "missing assignment id", edit: func(record *AssignmentOperationRecord) { record.AssignmentID = " " }, want: "assignment operation assignment_id is required"},
		{name: "missing agent id", edit: func(record *AssignmentOperationRecord) { record.AgentID = " " }, want: "assignment operation agent_id is required"},
		{name: "assignment id mismatch", edit: func(record *AssignmentOperationRecord) { record.Assignment.ID = "asn-other" }, want: "assignment operation assignment_id mismatch"},
		{name: "task id mismatch", edit: func(record *AssignmentOperationRecord) { record.Assignment.TaskID = "task-other" }, want: "assignment operation task_id mismatch"},
		{name: "agent id mismatch", edit: func(record *AssignmentOperationRecord) { record.Assignment.AgentID = "agent-other" }, want: "assignment operation agent_id mismatch"},
		{name: "missing events", edit: func(record *AssignmentOperationRecord) { record.Events = nil }, want: "assignment operation events are required"},
		{name: "non positive event seq", edit: func(record *AssignmentOperationRecord) { record.Events[0].Seq = 0 }, want: "assignment operation event seq must be positive"},
		{name: "event task mismatch", edit: func(record *AssignmentOperationRecord) { record.Events[0].TaskID = "task-other" }, want: "assignment operation event task_id mismatch"},
		{name: "missing recorded at", edit: func(record *AssignmentOperationRecord) { record.RecordedAt = time.Time{} }, want: "assignment operation recorded_at is required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			record := base
			record.Events = append([]TaskEvent(nil), base.Events...)
			if tc.edit != nil {
				tc.edit(&record)
			}
			err := validateAssignmentOperationRecord(record)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("validateAssignmentOperationRecord() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("validateAssignmentOperationRecord() error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestAssignmentOperationIDAndJSONShape(t *testing.T) {
	now := time.Date(2026, 5, 27, 13, 0, 0, 0, time.UTC)
	record := validAssignmentOperationRecord(now)
	if got, want := assignmentOperationID(AssignmentOperationAssignTask, record.Assignment, record.Events), "assign:asn-000042:7"; got != want {
		t.Fatalf("assign operation id = %q, want %q", got, want)
	}
	if got, want := assignmentOperationID(AssignmentOperationPollStart, record.Assignment, record.Events), "poll-start:asn-000042:lease-1:7"; got != want {
		t.Fatalf("poll-start operation id = %q, want %q", got, want)
	}
	if got, want := assignmentOperationID(AssignmentOperationAgentEvent, record.Assignment, record.Events), "agent-event:asn-000042:7"; got != want {
		t.Fatalf("agent-event operation id = %q, want %q", got, want)
	}
	if got, want := assignmentOperationID(AssignmentOperationClientStop, record.Assignment, record.Events), "client-stop:asn-000042:7"; got != want {
		t.Fatalf("client-stop operation id = %q, want %q", got, want)
	}
	if got, want := assignmentOperationID(AssignmentOperationType("custom"), record.Assignment, nil), "custom:asn-000042:0"; got != want {
		t.Fatalf("custom operation id = %q, want %q", got, want)
	}

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal assignment operation record: %v", err)
	}
	want := `{"schema_version":"riido-ai-server-assignment-operation.v1","operation_id":"assign:asn-000042:7","operation_type":"assign_task","task_id":"task-a","assignment_id":"asn-000042","agent_id":"agent-a","assignment":{"assignment_id":"asn-000042","task_id":"task-a","component_id":"component-a","agent_id":"agent-a","runtime_provider":"codex","prompt":"run tests","state":"queued","lease_token":"lease-1","created_at":"2026-05-27T13:00:00Z","updated_at":"2026-05-27T13:00:00Z"},"events":[{"seq":7,"task_id":"task-a","assignment_id":"asn-000042","agent_id":"agent-a","type":"assignment_queued","state":"queued","message":"queued","at":"2026-05-27T13:00:00Z"}],"recorded_at":"2026-05-27T13:00:00Z"}`
	if got := string(data); got != want {
		t.Fatalf("assignment operation JSON = %s, want %s", got, want)
	}
}

func TestAssignmentActiveLeaseExpired(t *testing.T) {
	now := time.Date(2026, 5, 27, 13, 0, 0, 0, time.UTC)
	cases := []struct {
		name  string
		lease AssignmentActiveLease
		want  bool
	}{
		{name: "unix ms future", lease: AssignmentActiveLease{LeaseExpiresUnixMS: now.Add(time.Second).UnixMilli(), LeaseExpiresAt: now.Add(-time.Hour)}, want: false},
		{name: "unix ms expired takes precedence", lease: AssignmentActiveLease{LeaseExpiresUnixMS: now.UnixMilli(), LeaseExpiresAt: now.Add(time.Hour)}, want: true},
		{name: "time future", lease: AssignmentActiveLease{LeaseExpiresAt: now.Add(time.Second)}, want: false},
		{name: "time expired at boundary", lease: AssignmentActiveLease{LeaseExpiresAt: now}, want: true},
		{name: "empty lease expired", lease: AssignmentActiveLease{}, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.lease.Expired(now); got != tc.want {
				t.Fatalf("Expired() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAssignmentOperationHelpers(t *testing.T) {
	now := time.Date(2026, 5, 27, 13, 0, 0, 0, time.UTC)
	record := validAssignmentOperationRecord(now)
	record.Events = append(record.Events, TaskEvent{Seq: 12, TaskID: "task-a"})
	if got, want := assignmentOperationLastEventSeq(record), int64(12); got != want {
		t.Fatalf("assignmentOperationLastEventSeq = %d, want %d", got, want)
	}
	if got := assignmentOperationLastEventSeq(AssignmentOperationRecord{}); got != 0 {
		t.Fatalf("empty assignmentOperationLastEventSeq = %d, want 0", got)
	}

	for id, want := range map[string]int64{"asn-000042": 42, " asn-7 ": 7, "task-1": 0, "asn-not-number": 0} {
		if got := assignmentSequence(id); got != want {
			t.Fatalf("assignmentSequence(%q) = %d, want %d", id, got, want)
		}
	}

	assignment := Assignment{ID: "asn-000042", CreatedAt: now}
	if got, want := assignmentQueueSort(assignment), "20260527T130000.000000000Z#asn-000042"; got != want {
		t.Fatalf("assignmentQueueSort = %q, want %q", got, want)
	}
	assignment = Assignment{ID: "asn-000043", UpdatedAt: now.Add(time.Minute)}
	if got, want := assignmentQueueSort(assignment), "20260527T130100.000000000Z#asn-000043"; got != want {
		t.Fatalf("assignmentQueueSort updated fallback = %q, want %q", got, want)
	}
}

func TestAssignmentOperationPortsCompile(t *testing.T) {
	fake := &fakeAssignmentOperationPort{}
	var _ AssignmentOperationStore = fake
	var _ AssignmentOperationLoader = fake
	var _ AssignmentQueueReader = fake
	var _ AssignmentClaimer = fake
	var _ AssignmentActiveLeaseStore = fake
	var _ AssignmentProjectionReader = fake

	if err := fake.SaveAssignmentOperation(context.Background(), validAssignmentOperationRecord(time.Now().UTC())); err != nil {
		t.Fatalf("SaveAssignmentOperation: %v", err)
	}
}

func validAssignmentOperationRecord(now time.Time) AssignmentOperationRecord {
	assignment := Assignment{
		ID:              "asn-000042",
		TaskID:          "task-a",
		ComponentID:     "component-a",
		AgentID:         "agent-a",
		RuntimeProvider: "codex",
		Prompt:          "run tests",
		State:           AssignmentQueued,
		LeaseToken:      "lease-1",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	events := []TaskEvent{{
		Seq:          7,
		TaskID:       assignment.TaskID,
		AssignmentID: assignment.ID,
		AgentID:      assignment.AgentID,
		Type:         EventAssignmentQueued,
		State:        AssignmentQueued,
		Message:      "queued",
		At:           now,
	}}
	return AssignmentOperationRecord{
		SchemaVersion: AssignmentOperationSchemaVersion,
		OperationID:   assignmentOperationID(AssignmentOperationAssignTask, assignment, events),
		OperationType: AssignmentOperationAssignTask,
		TaskID:        assignment.TaskID,
		AssignmentID:  assignment.ID,
		AgentID:       assignment.AgentID,
		Assignment:    assignment,
		Events:        events,
		RecordedAt:    now,
	}
}

type fakeAssignmentOperationPort struct{}

func (fakeAssignmentOperationPort) SaveAssignmentOperation(context.Context, AssignmentOperationRecord) error {
	return nil
}

func (fakeAssignmentOperationPort) LoadAssignmentOperations(context.Context) ([]AssignmentOperationRecord, error) {
	return nil, nil
}

func (fakeAssignmentOperationPort) LoadAgentQueueAssignments(context.Context, string) ([]Assignment, error) {
	return nil, nil
}

func (fakeAssignmentOperationPort) ClaimNextAssignment(context.Context, string, time.Time) (AssignmentClaimResult, error) {
	return AssignmentClaimResult{}, nil
}

func (fakeAssignmentOperationPort) LoadAgentActiveAssignment(context.Context, string) (AssignmentActiveLease, bool, error) {
	return AssignmentActiveLease{}, false, nil
}

func (fakeAssignmentOperationPort) RefreshAgentActiveAssignment(context.Context, Assignment, time.Time) error {
	return nil
}

func (fakeAssignmentOperationPort) LoadAssignmentProjection(context.Context, string) (AssignmentProjection, bool, error) {
	return AssignmentProjection{}, false, nil
}

func (fakeAssignmentOperationPort) Close() error {
	return nil
}
