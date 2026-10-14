package riidoaiserver

import (
	"testing"
	"time"
)

func TestRepairQueuedAssignmentBlockerClearsMissingBlocker(t *testing.T) {
	now := time.Date(2026, 7, 7, 8, 0, 0, 0, time.UTC)
	store := blockerRepairStore(now, nil)
	state := newStoreState()
	assignment := blockerRepairAssignment("asn-current", "asn-missing", AssignmentQueued, now)
	waiter := make(chan struct{}, 1)
	state.agentWaiters["agent-new"] = map[int64]chan struct{}{1: waiter}

	if err := store.repairQueuedAssignmentBlockerForClaim(&state, &assignment); err != nil {
		t.Fatalf("repair: %v", err)
	}

	assertBlockerRepairCurrent(t, state, "asn-current", "")
	assertBlockerRepairEvent(t, state.events["task-a"], 0, EventAssignmentQueued, "missing blocker cleared before daemon lease")
	if state.events["task-a"][0].Metadata["blocked_by_assignment_id"] != "asn-missing" {
		t.Fatalf("metadata = %+v", state.events["task-a"][0].Metadata)
	}
	select {
	case <-waiter:
	default:
		t.Fatal("waiter was not signaled")
	}
}

func TestRepairQueuedAssignmentBlockerCancelsQueuedBlocker(t *testing.T) {
	now := time.Date(2026, 7, 7, 8, 5, 0, 0, time.UTC)
	store := blockerRepairStore(now, nil)
	state := newStoreState()
	blocker := blockerRepairAssignment("asn-blocker", "", AssignmentQueued, now)
	blocker.AgentID = "agent-old"
	state.assignments[blocker.ID] = blocker
	assignment := blockerRepairAssignment("asn-current", blocker.ID, AssignmentQueued, now)

	if err := store.repairQueuedAssignmentBlockerForClaim(&state, &assignment); err != nil {
		t.Fatalf("repair: %v", err)
	}

	if got := state.assignments[blocker.ID]; got.State != AssignmentCancelled {
		t.Fatalf("blocker = %+v, want cancelled", got)
	}
	assertBlockerRepairCurrent(t, state, "asn-current", "")
	assertBlockerRepairEvent(t, state.events["task-a"], 0, EventAssignmentCancelled, "queued blocker was cancelled before queued assignment claim")
	assertBlockerRepairEvent(t, state.events["task-a"], 1, EventAssignmentQueued, "queued blocker cleared before daemon lease")
}

func TestRepairQueuedAssignmentBlockerFailsStaleActiveBlocker(t *testing.T) {
	now := time.Date(2026, 7, 7, 8, 10, 0, 0, time.UTC)
	operationStore := &runtimeFakeActiveLeaseOperationStore{activeFound: false}
	store := blockerRepairStore(now, operationStore)
	state := newStoreState()
	blocker := blockerRepairAssignment("asn-blocker", "", AssignmentRunning, now.Add(-time.Hour))
	blocker.AgentID = "agent-old"
	state.assignments[blocker.ID] = blocker
	assignment := blockerRepairAssignment("asn-current", blocker.ID, AssignmentQueued, now)

	if err := store.repairQueuedAssignmentBlockerForClaim(&state, &assignment); err != nil {
		t.Fatalf("repair: %v", err)
	}

	if got := state.assignments[blocker.ID]; got.State != AssignmentFailed {
		t.Fatalf("blocker = %+v, want failed", got)
	}
	assertBlockerRepairCurrent(t, state, "asn-current", "")
	assertBlockerRepairEvent(t, state.events["task-a"], 0, EventAssignmentFailed, "blocked queued assignment repaired after stale blocker lease expired")
	assertBlockerRepairEvent(t, state.events["task-a"], 1, EventAssignmentQueued, "stale blocker cleared before daemon lease")
}
