package riidoaiserver

import "testing"

func TestAssignmentForClientStopPrefersExplicitAssignment(t *testing.T) {
	state := newStoreState()
	state.assignments["asn-explicit"] = Assignment{
		ID: "asn-explicit", TaskID: "other-task", AgentID: "agent-a",
		State: AssignmentCompleted,
	}

	got, ok := state.assignmentForClientStop("task-a", CancelAssignmentRequest{
		AgentID: "agent-a", AssignmentID: "asn-explicit",
	})
	if !ok || got.ID != "asn-explicit" {
		t.Fatalf("explicit stop target = %+v ok=%v", got, ok)
	}
}

func TestAssignmentForClientStopSkipsTerminalAndForeignTask(t *testing.T) {
	state := newStoreState()
	state.assignments["asn-done"] = Assignment{
		ID: "asn-done", TaskID: "task-a", AgentID: "agent-a",
		State: AssignmentCompleted,
	}
	state.assignments["asn-foreign"] = Assignment{
		ID: "asn-foreign", TaskID: "task-b", AgentID: "agent-a",
		State: AssignmentRunning,
	}
	state.assignments["asn-active"] = Assignment{
		ID: "asn-active", TaskID: "task-a", AgentID: "agent-a",
		State: AssignmentRunning,
	}
	state.agentAssignments["agent-a"] = []string{"asn-done", "asn-foreign", "asn-active"}

	got, ok := state.assignmentForClientStop("task-a", CancelAssignmentRequest{AgentID: "agent-a"})
	if !ok || got.ID != "asn-active" {
		t.Fatalf("derived stop target = %+v ok=%v", got, ok)
	}
}

func TestAssignmentBlockerClearedOnlyWhenTerminal(t *testing.T) {
	state := newStoreState()
	blocked := Assignment{ID: "asn-queued", BlockedByAssignmentID: "asn-blocker"}
	state.assignments["asn-blocker"] = Assignment{ID: "asn-blocker", State: AssignmentRunning}
	if assignmentBlockerCleared(&state, blocked) {
		t.Fatal("running blocker should not be cleared")
	}
	state.assignments["asn-blocker"] = Assignment{ID: "asn-blocker", State: AssignmentFailed}
	if !assignmentBlockerCleared(&state, blocked) {
		t.Fatal("terminal blocker should be cleared")
	}
}
