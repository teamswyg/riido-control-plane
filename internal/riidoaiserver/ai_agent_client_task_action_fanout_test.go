package riidoaiserver

import (
	"testing"
	"time"
)

func TestFailureDiagnosticsEqual(t *testing.T) {
	base := &AIAgentTaskThreadFailureDiagnostics{ResultStatus: "failed", FailureCategory: "tool", Message: "denied"}
	same := &AIAgentTaskThreadFailureDiagnostics{ResultStatus: "failed", FailureCategory: "tool", Message: "denied"}
	diff := &AIAgentTaskThreadFailureDiagnostics{ResultStatus: "failed", FailureCategory: "tool", Message: "timeout"}
	if !failureDiagnosticsEqual(nil, nil) {
		t.Fatal("nil diagnostics should be equal")
	}
	if failureDiagnosticsEqual(base, nil) {
		t.Fatal("diagnostic and nil should differ")
	}
	if !failureDiagnosticsEqual(base, same) {
		t.Fatal("matching diagnostics should be equal")
	}
	if failureDiagnosticsEqual(base, diff) {
		t.Fatal("different diagnostics should differ")
	}
}

func TestQueueDiagnosticsEqual(t *testing.T) {
	now := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	base := &AIAgentTaskThreadQueueDiagnostics{
		Reason: "busy", BlockedByAssignmentID: "asn-1", BlockerAgentID: "agent-1",
		BlockerRuntimeProvider: "codex", BlockerState: AssignmentRunning, BlockerUpdatedAt: now,
	}
	same := *base
	diff := *base
	diff.BlockerUpdatedAt = now.Add(time.Second)
	if !queueDiagnosticsEqual(nil, nil) {
		t.Fatal("nil diagnostics should be equal")
	}
	if queueDiagnosticsEqual(base, nil) {
		t.Fatal("diagnostic and nil should differ")
	}
	if !queueDiagnosticsEqual(base, &same) {
		t.Fatal("matching diagnostics should be equal")
	}
	if queueDiagnosticsEqual(base, &diff) {
		t.Fatal("different diagnostics should differ")
	}
}

func TestShouldFanoutAgentTaskActionEvent(t *testing.T) {
	previous := AIAgentTaskThreadRecord{
		WorkStatus: AgentWorkStatusRunning, AssignmentState: AgentAssignmentStateRunning,
		CommentKind:        AgentTaskCommentAssignmentStarted,
		FailureDiagnostics: &AIAgentTaskThreadFailureDiagnostics{Message: "old"},
	}
	response := AIAgentTaskActionResponse{
		WorkStatus: AgentWorkStatusRunning, AssignmentState: AgentAssignmentStateRunning,
		CommentKind:        AgentTaskCommentAssignmentStarted,
		FailureDiagnostics: &AIAgentTaskThreadFailureDiagnostics{Message: "old"},
	}
	if !shouldFanoutAgentTaskActionEvent(false, previous, response) {
		t.Fatal("new thread should fan out")
	}
	if shouldFanoutAgentTaskActionEvent(true, previous, response) {
		t.Fatal("unchanged thread should not fan out")
	}
	response.CommentKind = AgentTaskCommentTaskFailed
	if !shouldFanoutAgentTaskActionEvent(true, previous, response) {
		t.Fatal("comment kind change should fan out")
	}
	response.CommentKind = previous.CommentKind
	response.FailureDiagnostics = &AIAgentTaskThreadFailureDiagnostics{Message: "new"}
	if !shouldFanoutAgentTaskActionEvent(true, previous, response) {
		t.Fatal("failure diagnostic change should fan out")
	}
}
