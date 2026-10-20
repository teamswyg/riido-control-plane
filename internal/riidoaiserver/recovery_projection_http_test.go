package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestRecoveryHTTPKeepsJournalReplayAndTerminal(t *testing.T) {
	f := newRecoveryProjectionHTTPFixture(t)
	before := f.snapshots.saves
	f.snapshots.fail = true
	first := f.post(t, f.recovery, http.StatusOK)
	replay := f.post(t, f.recovery, http.StatusOK)
	if len(f.ops.records) != 1 || first.Event.Seq != replay.Event.Seq || f.snapshots.saves != before {
		t.Fatal("recovery journal replay or zero snapshot I/O broken")
	}
	f.recovery.Metadata["lease_token"] = "wrong-lease"
	f.post(t, f.recovery, http.StatusForbidden)
	if len(f.ops.records) != 1 {
		t.Fatal("invalid lease bypassed journal fence")
	}
	terminal := AgentEventRequest{AssignmentID: f.current.ID, TaskID: f.current.TaskID, DaemonID: f.recovery.DaemonID, DeviceID: f.recovery.DeviceID, RuntimeID: f.recovery.RuntimeID, State: AssignmentCompleted, EventType: EventAssignmentCompleted, Message: "fixture answer"}
	f.post(t, terminal, http.StatusServiceUnavailable)
	if len(f.ops.records) != 2 || f.ops.projection.Assignment.State != AssignmentCompleted {
		t.Fatal("terminal not journaled before failed client projection")
	}
	f.snapshots.fail = false
	accepted := f.post(t, terminal, http.StatusOK)
	if len(f.ops.records) != 2 || accepted.Assignment.State != AssignmentCompleted || f.snapshots.saves <= before {
		t.Fatal("same terminal retry did not repair durable client snapshot")
	}
	threads := f.snapshots.snapshot.TaskThreads[f.current.TaskID]
	if len(threads) != 1 || threads[0].AssignmentState != AgentAssignmentStateCompleted {
		t.Fatal("saved terminal history lost after recovery-only no-op")
	}
}
