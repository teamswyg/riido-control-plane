package riidoaiserver

import (
	"strings"
	"testing"
	"time"
)

func TestAssignmentOperationReplayKeepsSameEventSeqAcrossDifferentTasks(t *testing.T) {
	fixedNow := time.Date(2026, 5, 27, 16, 0, 0, 0, time.UTC)
	state, err := stateFromAssignmentOperations([]AssignmentOperationRecord{
		replayOperationRecord("task-a", "asn-000001", "agent-a", AssignmentQueued, 1, fixedNow),
		replayOperationRecord("task-b", "asn-000002", "agent-a", AssignmentQueued, 1, fixedNow),
	})
	if err != nil {
		t.Fatalf("stateFromAssignmentOperations: %v", err)
	}
	if len(state.events["task-a"]) != 1 || len(state.events["task-b"]) != 1 {
		t.Fatalf("events = %+v", state.events)
	}
	if state.nextEventSeq != 1 || state.nextAssignmentSeq != 2 {
		t.Fatalf("state seqs nextEvent=%d nextAssignment=%d", state.nextEventSeq, state.nextAssignmentSeq)
	}
}

func TestAssignmentOperationReplayOrdersAndDedupesEvents(t *testing.T) {
	base := time.Date(2026, 5, 27, 16, 0, 0, 0, time.UTC)
	first := replayOperationRecord("task-a", "asn-000001", "agent-a", AssignmentQueued, 1, base)
	second := replayOperationRecord("task-a", "asn-000001", "agent-a", AssignmentRunning, 2, base.Add(time.Minute))
	duplicate := first
	duplicate.OperationID = "duplicate-late"
	duplicate.RecordedAt = base.Add(2 * time.Minute)

	state, err := stateFromAssignmentOperations([]AssignmentOperationRecord{second, duplicate, first})
	if err != nil {
		t.Fatalf("stateFromAssignmentOperations: %v", err)
	}
	events := state.events["task-a"]
	if len(events) != 2 {
		t.Fatalf("events = %+v", events)
	}
	if events[0].Seq != 1 || events[1].Seq != 2 {
		t.Fatalf("events not sorted: %+v", events)
	}
	if got := state.assignments["asn-000001"].State; got != AssignmentRunning {
		t.Fatalf("assignment state = %q, want %q", got, AssignmentRunning)
	}
}

func TestAssignmentOperationReplayRebuildsAssignmentIndexes(t *testing.T) {
	base := time.Date(2026, 5, 27, 16, 0, 0, 0, time.UTC)
	oldAssignment := replayOperationRecord("task-a", "asn-000001", "agent-a", AssignmentCancelling, 1, base)
	newAssignment := replayOperationRecord("task-a", "asn-000002", "agent-b", AssignmentQueued, 2, base.Add(time.Minute))
	state, err := stateFromAssignmentOperations([]AssignmentOperationRecord{newAssignment, oldAssignment})
	if err != nil {
		t.Fatalf("stateFromAssignmentOperations: %v", err)
	}
	task := state.tasks["task-a"]
	if task.currentAssignmentID != "asn-000002" || task.componentID != "component-task-a" {
		t.Fatalf("task index = %+v", task)
	}
	if got := state.agentAssignments["agent-a"]; len(got) != 1 || got[0] != "asn-000001" {
		t.Fatalf("agent-a assignments = %v", got)
	}
	if got := state.agentAssignments["agent-b"]; len(got) != 1 || got[0] != "asn-000002" {
		t.Fatalf("agent-b assignments = %v", got)
	}
}

func TestAssignmentOperationReplayRejectsInvalidRecords(t *testing.T) {
	record := replayOperationRecord("task-a", "asn-000001", "agent-a", AssignmentQueued, 1, time.Date(2026, 5, 27, 16, 0, 0, 0, time.UTC))
	record.SchemaVersion = "riido-ai-server-assignment-operation.v0"
	_, err := stateFromAssignmentOperations([]AssignmentOperationRecord{record})
	if err == nil || !strings.Contains(err.Error(), "unsupported assignment operation schema_version") {
		t.Fatalf("replay invalid record err=%v", err)
	}
}

func replayOperationRecord(taskID, assignmentID, agentID string, state AssignmentState, eventSeq int64, at time.Time) AssignmentOperationRecord {
	assignment := Assignment{
		ID:              assignmentID,
		TaskID:          taskID,
		ComponentID:     "component-" + taskID,
		AgentID:         agentID,
		RuntimeProvider: "codex",
		Prompt:          "run migration test",
		State:           state,
		CreatedAt:       at,
		UpdatedAt:       at,
	}
	events := []TaskEvent{{
		Seq:          eventSeq,
		TaskID:       taskID,
		AssignmentID: assignmentID,
		AgentID:      agentID,
		Type:         EventAssignmentStateUpdated,
		State:        state,
		At:           at,
	}}
	operationType := AssignmentOperationAgentEvent
	if eventSeq == 1 {
		operationType = AssignmentOperationAssignTask
		events[0].Type = EventAssignmentQueued
	}
	return AssignmentOperationRecord{
		SchemaVersion: AssignmentOperationSchemaVersion,
		OperationID:   assignmentOperationID(operationType, assignment, events),
		OperationType: operationType,
		TaskID:        taskID,
		AssignmentID:  assignmentID,
		AgentID:       agentID,
		Assignment:    assignment,
		Events:        events,
		RecordedAt:    at,
	}
}
