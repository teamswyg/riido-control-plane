package riidoaiserver

import (
	"testing"
	"time"
)

func blockerRepairStore(now time.Time, operationStore AssignmentOperationStore) *Store {
	return &Store{
		now:                 func() time.Time { return now },
		activeLeaseDuration: time.Second,
		operationStore:      operationStore,
		operationMetrics:    NewStoreOperationMetrics(),
	}
}

func blockerRepairAssignment(id, blockerID string, state AssignmentState, updatedAt time.Time) Assignment {
	return Assignment{
		ID: id, TaskID: "task-a", AgentID: "agent-new",
		RuntimeProvider: "codex", Prompt: "run",
		State:                 blockerRepairState(state),
		BlockedByAssignmentID: blockerID,
		UpdatedAt:             updatedAt,
		LeaseToken:            "lease-" + id,
	}
}

func blockerRepairState(state AssignmentState) AssignmentState {
	return state
}

func assertBlockerRepairCurrent(t *testing.T, state storeState, id, blockerID string) {
	t.Helper()
	if got := state.assignments[id]; got.BlockedByAssignmentID != blockerID {
		t.Fatalf("assignment = %+v, want blocker %q", got, blockerID)
	}
}

func assertBlockerRepairEvent(t *testing.T, events []TaskEvent, index int, eventType, message string) {
	t.Helper()
	if len(events) <= index {
		t.Fatalf("events = %+v, want index %d", events, index)
	}
	if event := events[index]; event.Type != eventType || event.Message != message {
		t.Fatalf("event[%d] = %+v", index, event)
	}
}
