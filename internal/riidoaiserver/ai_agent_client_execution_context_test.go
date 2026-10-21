package riidoaiserver

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
)

func TestExecutionContextSurvivesTaskHistoryAndEventProjections(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	ctx := context.Background()
	principal := AuthorizationResult{PrincipalID: "user-1"}
	taskID := "task-execution-context"
	assigned, err := store.AssignAIAgentTask(ctx, principal, taskID, AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", AssignmentID: "asn-execution-context"})
	if err != nil {
		t.Fatal(err)
	}
	want := assignment.ExecutionContext{WorkspaceKind: "repository_clone", VerificationStatus: "failed", DiagnosticCode: "host_verification_failed", NextAction: "review_result"}
	event := TaskEvent{TaskID: taskID, AssignmentID: assigned.AssignmentID, AgentID: assigned.AgentID, Type: EventAssignmentFailed, State: AssignmentFailed, Metadata: want.Metadata(), At: time.Now().UTC()}
	if err := store.RecordAIAgentAssignmentEvent(ctx, assigned.AgentID, AgentEventRequest{}, event); err != nil {
		t.Fatal(err)
	}
	threads, err := store.ListAIAgentTaskThreads(ctx, principal, taskID)
	if err != nil || len(threads.Threads) != 1 {
		t.Fatalf("threads: %v %+v", err, threads)
	}
	assertContext := func(label string, got *assignment.ExecutionContext) {
		t.Helper()
		if got == nil || *got != want {
			t.Fatalf("%s: got %+v, want %+v", label, got, want)
		}
	}
	thread := threads.Threads[0]
	assertContext("v1 cold read", thread.ExecutionContext)
	assertContext("action response", actionResponseFromThread(thread, "").ExecutionContext)
	history, err := store.ListAIAgentTaskThreadHistory(ctx, principal, taskID)
	if err != nil || len(history.Threads) != 1 {
		t.Fatalf("history: %v", err)
	}
	assertContext("v3 cold history", history.Threads[0].ExecutionContext)
	notification, ok := lastWorkStatusChangedEventForTask(t, store, principal, taskID)
	if !ok {
		t.Fatal("missing SSE event")
	}
	assertContext("SSE", notification.ExecutionContext)
	// Persistence serializes records; a reload must preserve typed evidence.
	data, err := json.Marshal(thread)
	if err != nil {
		t.Fatal(err)
	}
	var restored AIAgentTaskThreadRecord
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	assertContext("persisted record", restored.ExecutionContext)
	copied := copyTaskThread(restored)
	copied.ExecutionContext.DiagnosticCode = "unknown"
	assertContext("copy does not alias stored record", restored.ExecutionContext)
}

func TestExecutionContextRetainsLegacyProgressAndRejectsUnboundedMetadata(t *testing.T) {
	original := assignment.ExecutionContext{WorkspaceKind: "repository_clone", VerificationStatus: "pending", DiagnosticCode: "none", NextAction: "none"}
	thread := AIAgentTaskThreadRecord{ExecutionContext: &original}
	// Event streams from older daemons omit the new metadata entirely.
	response := assignmentEventActionResponse(thread, AssignmentRunning, "", nil)
	if response.ExecutionContext == nil || *response.ExecutionContext != original {
		t.Fatal("legacy progress erased existing evidence")
	}
	unsafe := assignment.ExecutionContext{WorkspaceKind: "/private/customer", VerificationStatus: "passed: secret", DiagnosticCode: "private command", NextAction: "run secret"}
	result := copyExecutionContext(&unsafe)
	body, _ := json.Marshal(result)
	if strings.Contains(string(body), "private") || strings.Contains(string(body), "secret") {
		t.Fatal("unbounded diagnostic escaped")
	}
	if result.WorkspaceKind != "unknown" || result.VerificationStatus != "unknown" {
		t.Fatal("unknown evidence was presented as verified")
	}
	if copyExecutionContext(nil) != nil {
		t.Fatal("legacy record invented evidence")
	}
	changed := original
	changed.VerificationStatus = "passed"
	if !shouldFanoutAgentTaskActionEvent(true, thread, AIAgentTaskActionResponse{ExecutionContext: &changed}) {
		t.Fatal("execution context transition did not fan out")
	}
}

func TestNewRunDoesNotInheritExecutionEvidence(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	evidence := assignment.ExecutionContext{WorkspaceKind: "repository_clone", VerificationStatus: "passed", DiagnosticCode: "none", NextAction: "none"}
	thread := AIAgentTaskThreadRecord{ThreadID: "thread-a", TaskID: "task-a", AssignmentID: "old", RunID: "run-old", ExecutionContext: &evidence}
	store.updateTaskThreadMessageFromActionLocked(&thread, AIAgentTaskActionResponse{ThreadID: "thread-a", TaskID: "task-a", AssignmentID: "new", RunID: "run-new", AssignmentState: AgentAssignmentStateQueued}, "", "", "", time.Now().UTC())
	if thread.ExecutionContext != nil {
		t.Fatal("new run inherited prior host verification")
	}
	if thread.AssignmentID != "new" || thread.RunID != "run-new" {
		t.Fatal("run identity was not updated")
	}
}
