package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

// A deleted agent's ID is reused by the next create with the same owner and runtime.
// The new agent must not inherit the old provenance or the one-time completion record.
func TestOnboardingDeletedAgentIDReuseDoesNotInheritCompletion(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	store := NewDevelopmentAIAgentClientStore()
	const firstTask, secondTask = "task-reuse-1", "task-reuse-2"

	fixture, err := store.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("reuse"))
	if err != nil {
		t.Fatalf("fixture create: %v", err)
	}
	agentID := fixture.Agent.AgentID
	request := func() CompleteAIAgentOnboardingTaskRequest {
		return CompleteAIAgentOnboardingTaskRequest{AgentID: agentID, TemplateKey: AIAgentOnboardingTemplateSchedule}
	}
	if _, err := store.CompleteAIAgentOnboardingTask(ctx, principal, firstTask, request()); err != nil {
		t.Fatalf("first completion: %v", err)
	}
	if _, err := store.DeleteAIAgent(ctx, principal, agentID); err != nil {
		t.Fatalf("delete fixture agent: %v", err)
	}

	normal, err := store.CreateAIAgent(ctx, principal, profileColorTestRequest("reuse"))
	if err != nil {
		t.Fatalf("normal create: %v", err)
	}
	if normal.Agent.AgentID != agentID {
		t.Fatalf("precondition: deleted agent ID %q was not reused, got %q", agentID, normal.Agent.AgentID)
	}
	_, err = store.CompleteAIAgentOnboardingTask(ctx, principal, secondTask, request())
	if !errors.Is(err, errOnboardingCompletionIneligible) {
		t.Fatalf("normal agent with reused ID completion error = %v, want ineligible", err)
	}
	if _, err := store.DeleteAIAgent(ctx, principal, agentID); err != nil {
		t.Fatalf("delete normal agent: %v", err)
	}

	recreated, err := store.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("reuse"))
	if err != nil {
		t.Fatalf("fixture recreate: %v", err)
	}
	if recreated.Agent.AgentID != agentID {
		t.Fatalf("precondition: deleted agent ID %q was not reused, got %q", agentID, recreated.Agent.AgentID)
	}
	if _, err := store.CompleteAIAgentOnboardingTask(ctx, principal, secondTask, request()); err != nil {
		t.Fatalf("fixture-created agent completion on a new task: %v", err)
	}
	// The earlier completion must not replay for the new agent on the old task.
	_, err = store.CompleteAIAgentOnboardingTask(ctx, principal, firstTask, request())
	if !errors.Is(err, errOnboardingCompletionConflict) {
		t.Fatalf("old task completion for reused ID error = %v, want conflict", err)
	}
}
