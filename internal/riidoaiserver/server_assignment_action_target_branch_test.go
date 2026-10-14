package riidoaiserver

import (
	"strings"
	"testing"
)

func TestActiveAssignmentActionTargetUsesLatestRunning(t *testing.T) {
	threads := []AIAgentTaskThreadRecord{
		actionTargetThread("agent-a", "asn-old", AgentAssignmentStateRunning),
		actionTargetThread("agent-b", "asn-new", AgentAssignmentStateRunning),
	}
	target, ok, err := activeAssignmentActionTarget("task-a", threads)
	if err != nil || !ok {
		t.Fatalf("active target = %+v ok=%v err=%v", target, ok, err)
	}
	if target.AssignmentID != "asn-new" || target.AgentID != "agent-b" {
		t.Fatalf("target = %+v, want latest running assignment", target)
	}
}

func TestAssignmentActionTargetByAssignmentIDRejectsAgentMismatch(t *testing.T) {
	threads := []AIAgentTaskThreadRecord{
		actionTargetThread("agent-a", "asn-a", AgentAssignmentStateRunning),
	}
	_, ok, err := assignmentActionTargetByAssignmentID("task-a", "agent-b", "asn-a", threads)
	if ok || err == nil || !strings.Contains(err.Error(), "does not belong") {
		t.Fatalf("target mismatch = ok %v err %v", ok, err)
	}
}

func TestActionTargetFromThreadCanUseHistoricalAssignment(t *testing.T) {
	threads := []AIAgentTaskThreadRecord{
		actionTargetThread("agent-a", "asn-done", AgentAssignmentStateCompleted),
	}
	target, ok := actionTargetFromThread("task-a", "agent-a", threads, false)
	if !ok || target.AssignmentID != "asn-done" {
		t.Fatalf("historical target = %+v ok=%v", target, ok)
	}
}
