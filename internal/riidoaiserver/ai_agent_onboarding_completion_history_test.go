package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

// driveAssignmentRunState moves an assignment through the lifecycle the dashboard
// tests use, so the assignment store holds it as running or finished.
func driveAssignmentRunState(t *testing.T, store *Store, agentID, taskID, assignmentID string, finished bool) {
	t.Helper()
	ctx := context.Background()
	for _, state := range []AssignmentState{AssignmentLeased, AssignmentReady, AssignmentRunning} {
		if _, err := store.RecordAgentEvent(ctx, agentID, AgentEventRequest{AssignmentID: assignmentID, TaskID: taskID, DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-cursor-dev", State: state}); err != nil {
			t.Fatalf("record %s: %v", state, err)
		}
	}
	if !finished {
		return
	}
	if _, err := store.RecordAgentEvent(ctx, agentID, AgentEventRequest{AssignmentID: assignmentID, TaskID: taskID, DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-cursor-dev", EventType: EventAssignmentCompleted, State: AssignmentCompleted}); err != nil {
		t.Fatalf("record completed: %v", err)
	}
}

// A repeat of an existing completion returns the recorded response, even after a normal
// assignment on the same task has started running for this agent.
func TestHTTPOnboardingCompletionRepeatReturnsRecordedAfterLaterRun(t *testing.T) {
	handler, assignmentStore := newOnboardingCompletionHTTPTestServer(t)
	agentID := createOnboardingAgentHTTP(t, handler, "history repeat")
	const taskID = "task-history-repeat-1"
	body := onboardingCompletionBody(agentID, "schedule")
	first := decodeOnboardingCompletionResponse(t, postOnboardingCompletion(t, handler, taskID, body))

	assignment, err := assignmentStore.AssignTaskAdditive(t.Context(), taskID, AssignRequest{
		ComponentID:     taskID,
		AgentID:         agentID,
		RuntimeProvider: "cursor",
		Prompt:          "work",
	})
	if err != nil {
		t.Fatalf("later assignment: %v", err)
	}
	driveAssignmentRunState(t, assignmentStore, agentID, taskID, assignment.ID, false)

	repeat := decodeOnboardingCompletionResponse(t, postOnboardingCompletion(t, handler, taskID, body))
	if repeat.AssignmentID != first.AssignmentID || repeat.ThreadID != first.ThreadID || repeat.Message != first.Message {
		t.Fatalf("repeat = %+v, want recorded %+v", repeat, first)
	}
}

// A finished run on the task with no thread in the client store is history, so a new
// completion is refused.
func TestHTTPOnboardingCompletionRefusesFinishedRunWithoutThread(t *testing.T) {
	handler, assignmentStore := newOnboardingCompletionHTTPTestServer(t)
	agentID := createOnboardingAgentHTTP(t, handler, "finished run")
	const taskID = "task-finished-run-1"
	assignment, err := assignmentStore.AssignTaskAdditive(t.Context(), taskID, AssignRequest{
		ComponentID:     taskID,
		AgentID:         agentID,
		RuntimeProvider: "cursor",
		Prompt:          "work",
	})
	if err != nil {
		t.Fatalf("assignment: %v", err)
	}
	driveAssignmentRunState(t, assignmentStore, agentID, taskID, assignment.ID, true)

	rec := postOnboardingCompletion(t, handler, taskID, onboardingCompletionBody(agentID, "schedule"))
	if rec.Code != http.StatusConflict {
		t.Fatalf("completion after finished run status=%d body=%s, want 409", rec.Code, rec.Body.String())
	}
}
