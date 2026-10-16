package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// A durable assignment that is not yet published as a thread must still block the
// completion. Without the gate and this check, the store sees no thread and accepts it.
func TestHTTPOnboardingCompletionRefusesDurableAssignmentWithoutThread(t *testing.T) {
	handler, assignmentStore := newOnboardingCompletionHTTPTestServer(t)
	busyAgentID := createOnboardingAgentHTTP(t, handler, "durable busy")
	idleAgentID := createOnboardingAgentHTTP(t, handler, "durable idle")
	const taskID = "task-durable-gap-1"
	if _, err := assignmentStore.AssignTaskAdditive(t.Context(), taskID, AssignRequest{
		ComponentID:     taskID,
		AgentID:         busyAgentID,
		RuntimeProvider: "cursor",
		Prompt:          "work",
	}); err != nil {
		t.Fatalf("durable assignment: %v", err)
	}

	busy := postOnboardingCompletion(t, handler, taskID, onboardingCompletionBody(busyAgentID, "schedule"))
	if busy.Code != http.StatusConflict {
		t.Fatalf("completion with durable assignment and no thread status=%d body=%s, want 409", busy.Code, busy.Body.String())
	}

	// Control: an agent with no durable assignment on the task still completes.
	idle := postOnboardingCompletion(t, handler, taskID, onboardingCompletionBody(idleAgentID, "schedule"))
	if idle.Code != http.StatusAccepted {
		t.Fatalf("control completion status=%d body=%s, want 202", idle.Code, idle.Body.String())
	}
}

// The completion must wait behind the task publication gate that an assignment holds
// from its durable write until its thread is published.
func TestOnboardingCompletionWaitsForTaskPublicationGate(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	client := NewDevelopmentAIAgentClientStore()
	assignments := NewStoreWithConfig(StoreConfig{AgentRegistry: client})
	t.Cleanup(assignments.Close)
	created, err := client.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("gate"))
	if err != nil {
		t.Fatalf("fixture create: %v", err)
	}
	server := Server{aiAgent: client, assignment: assignments, taskAgentExecutionGate: newTaskAgentExecutionGates()}
	const taskID = "task-gate-wait-1"
	release, err := server.taskAgentExecutionGate.beginPublication(ctx, taskID)
	if err != nil {
		t.Fatalf("hold publication gate: %v", err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := server.completeAIAgentOnboardingTaskUnderGate(ctx, principal, taskID, CompleteAIAgentOnboardingTaskRequest{
			AgentID:     created.Agent.AgentID,
			TemplateKey: AIAgentOnboardingTemplateSchedule,
		})
		done <- err
	}()
	select {
	case err := <-done:
		release()
		t.Fatalf("completion ran while the task publication gate was held: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	release()
	if err := <-done; err != nil {
		t.Fatalf("completion after gate release: %v", err)
	}
}
