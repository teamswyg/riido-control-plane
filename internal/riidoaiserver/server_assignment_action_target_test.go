package riidoaiserver

import "testing"

func TestAssignmentActionTargetPrefersRunningOverQueued(t *testing.T) {
	threads := []AIAgentTaskThreadRecord{
		actionTargetThread("agent-a", "asn-running", AgentAssignmentStateRunning),
		actionTargetThread("agent-a", "asn-queued", AgentAssignmentStateQueued),
	}

	target, ok := actionTargetFromThread("task-a", "agent-a", threads, true)
	if !ok {
		t.Fatal("target not found")
	}
	if target.AssignmentID != "asn-running" {
		t.Fatalf("target=%+v, want running assignment", target)
	}
}

func TestAssignmentActionTargetFallsBackToQueued(t *testing.T) {
	threads := []AIAgentTaskThreadRecord{
		actionTargetThread("agent-a", "asn-queued", AgentAssignmentStateQueued),
	}

	target, ok := actionTargetFromThread("task-a", "agent-a", threads, true)
	if !ok {
		t.Fatal("target not found")
	}
	if target.AssignmentID != "asn-queued" {
		t.Fatalf("target=%+v, want queued assignment", target)
	}
}

func TestAssignmentActionTargetsCollectRunningThenQueued(t *testing.T) {
	threads := []AIAgentTaskThreadRecord{
		actionTargetThread("agent-a", "asn-queued-old", AgentAssignmentStateQueued),
		actionTargetThread("agent-a", "asn-running", AgentAssignmentStateRunning),
		actionTargetThread("agent-a", "asn-queued-new", AgentAssignmentStateQueued),
		actionTargetThread("agent-b", "asn-other", AgentAssignmentStateRunning),
	}

	targets := activeAssignmentActionTargetsForAgent("task-a", "agent-a", threads)
	if len(targets) != 3 {
		t.Fatalf("targets=%+v, want 3", targets)
	}
	got := []string{targets[0].AssignmentID, targets[1].AssignmentID, targets[2].AssignmentID}
	want := []string{"asn-running", "asn-queued-new", "asn-queued-old"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("targets=%+v, want assignment order %v", targets, want)
		}
	}
}

func actionTargetThread(agentID, assignmentID string, state AgentAssignmentState) AIAgentTaskThreadRecord {
	return AIAgentTaskThreadRecord{
		ThreadID:        "thread-" + assignmentID,
		TaskID:          "task-a",
		AssignmentID:    assignmentID,
		AgentID:         agentID,
		RunID:           "run-" + assignmentID,
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: state,
	}
}
