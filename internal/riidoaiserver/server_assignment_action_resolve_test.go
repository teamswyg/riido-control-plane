package riidoaiserver

import (
	"strings"
	"testing"
)

func TestAssignmentActionTargetsFromThreadsSelectsByScope(t *testing.T) {
	threads := []AIAgentTaskThreadRecord{
		actionTargetThread("agent-a", "asn-done", AgentAssignmentStateCompleted),
		actionTargetThread("agent-a", "asn-queued", AgentAssignmentStateQueued),
		actionTargetThread("agent-a", "asn-running", AgentAssignmentStateRunning),
		actionTargetThread("agent-b", "asn-other", AgentAssignmentStateRunning),
	}

	targets, ok, err := assignmentActionTargetsFromThreads("task-a", "agent-a", "asn-queued", threads)
	if err != nil || !ok || len(targets) != 1 || targets[0].AssignmentID != "asn-queued" {
		t.Fatalf("assignment scoped targets=%+v ok=%v err=%v", targets, ok, err)
	}

	targets, ok, err = assignmentActionTargetsFromThreads("task-a", "agent-a", "", threads)
	if err != nil || !ok || len(targets) != 2 {
		t.Fatalf("agent scoped targets=%+v ok=%v err=%v", targets, ok, err)
	}
	if targets[0].AssignmentID != "asn-running" || targets[1].AssignmentID != "asn-queued" {
		t.Fatalf("agent scoped order=%+v", targets)
	}

	targets, ok, err = assignmentActionTargetsFromThreads("task-a", "", "", threads)
	if err != nil || !ok || len(targets) != 1 || targets[0].AssignmentID != "asn-other" {
		t.Fatalf("active scoped targets=%+v ok=%v err=%v", targets, ok, err)
	}
}

func TestAssignmentActionTargetsFromThreadsRejectsInactiveOrForeignTargets(t *testing.T) {
	threads := []AIAgentTaskThreadRecord{
		actionTargetThread("agent-a", "asn-done", AgentAssignmentStateCompleted),
		actionTargetThread("agent-a", "asn-current", AgentAssignmentStateRunning),
	}

	targets, ok, err := assignmentActionTargetsFromThreads("task-a", "agent-b", "asn-current", threads)
	if ok || err == nil || !strings.Contains(err.Error(), "does not belong") || targets != nil {
		t.Fatalf("foreign assignment targets=%+v ok=%v err=%v", targets, ok, err)
	}

	targets, ok, err = assignmentActionTargetsFromThreads("task-a", "", "asn-missing", threads)
	if ok || err == nil || !strings.Contains(err.Error(), "does not belong") || targets != nil {
		t.Fatalf("missing assignment targets=%+v ok=%v err=%v", targets, ok, err)
	}

	targets, ok, err = assignmentActionTargetsFromThreads("task-a", "agent-z", "", threads)
	if err != nil || ok || targets != nil {
		t.Fatalf("unknown agent targets=%+v ok=%v err=%v", targets, ok, err)
	}

	targets, ok, err = assignmentActionTargetsFromThreads("task-a", "", "", threads[:1])
	if err != nil || ok || targets != nil {
		t.Fatalf("terminal-only active target=%+v ok=%v err=%v", targets, ok, err)
	}
}
