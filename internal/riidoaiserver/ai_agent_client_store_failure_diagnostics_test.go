package riidoaiserver

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestDevelopmentAIAgentClientStoreExposesFailureDiagnostics(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1"}
	assigned, err := store.AssignAIAgentTask(context.Background(), principal, "task-failure-diagnostics", AssignAIAgentTaskRequest{
		AgentID:      "agent-owned-codex",
		AssignmentID: "asn-failure-diagnostics",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	failedEvent := TaskEvent{
		TaskID:       "task-failure-diagnostics",
		AssignmentID: assigned.AssignmentID,
		AgentID:      assigned.AgentID,
		Type:         EventAssignmentFailed,
		State:        AssignmentFailed,
		Message:      "approval_timeout: no headless approval path /tmp/riido/private.txt",
		Metadata: map[string]string{
			metadatakeys.AssignmentResultStatus.String():    "blocked",
			metadatakeys.AssignmentFailureCategory.String(): "provider_blocked",
		},
		At: time.Now().UTC(),
	}
	if err := store.RecordAIAgentAssignmentEvent(context.Background(), assigned.AgentID, AgentEventRequest{}, failedEvent); err != nil {
		t.Fatalf("RecordAIAgentAssignmentEvent: %v", err)
	}
	threads, err := store.ListAIAgentTaskThreads(context.Background(), principal, "task-failure-diagnostics")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	assertFailureDiagnostics(t, "thread", threads.Threads[0].FailureDiagnostics)
	response := actionResponseFromThread(threads.Threads[0], "")
	assertFailureDiagnostics(t, "action response", response.FailureDiagnostics)
	event, ok := lastWorkStatusChangedEventForTask(t, store, principal, "task-failure-diagnostics")
	if !ok {
		t.Fatal("missing work status changed event")
	}
	assertFailureDiagnostics(t, "work status event", event.FailureDiagnostics)
}
