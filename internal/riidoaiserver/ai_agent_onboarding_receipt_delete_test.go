package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// A request_id receipt of a deleted agent must not replay for the agent that reuses
// its ID. The reused agent has no onboarding provenance, so the call must be refused.
func TestOnboardingRequestReceiptDoesNotReplayForReusedAgentID(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	store := NewDevelopmentAIAgentClientStore()
	const taskID = "task-receipt-reuse-1"

	fixture, err := store.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("receipt"))
	if err != nil {
		t.Fatalf("fixture create: %v", err)
	}
	agentID := fixture.Agent.AgentID
	request := CompleteAIAgentOnboardingTaskRequest{
		AgentID:     agentID,
		TemplateKey: AIAgentOnboardingTemplateSchedule,
		RequestID:   "req-receipt-reuse-1",
	}
	if _, err := store.CompleteAIAgentOnboardingTask(ctx, principal, taskID, request); err != nil {
		t.Fatalf("first completion: %v", err)
	}
	if _, err := store.DeleteAIAgent(ctx, principal, agentID); err != nil {
		t.Fatalf("delete fixture agent: %v", err)
	}

	normal, err := store.CreateAIAgent(ctx, principal, profileColorTestRequest("receipt"))
	if err != nil {
		t.Fatalf("normal create: %v", err)
	}
	if normal.Agent.AgentID != agentID {
		t.Fatalf("precondition: deleted agent ID %q was not reused, got %q", agentID, normal.Agent.AgentID)
	}
	_, err = store.CompleteAIAgentOnboardingTask(ctx, principal, taskID, request)
	if !errors.Is(err, errOnboardingCompletionIneligible) {
		t.Fatalf("same request_id for reused agent ID error = %v, want ineligible (no replayed success)", err)
	}
}

// Deleting an agent removes its onboarding receipts from the persisted snapshot too,
// so a reload cannot bring them back.
func TestOnboardingRequestReceiptDeletedFromSnapshotWithAgent(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	store := NewDevelopmentAIAgentClientStore()
	fixture, err := store.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("receipt snapshot"))
	if err != nil {
		t.Fatalf("fixture create: %v", err)
	}
	agentID := fixture.Agent.AgentID
	if _, err := store.CompleteAIAgentOnboardingTask(ctx, principal, "task-receipt-snapshot-1", CompleteAIAgentOnboardingTaskRequest{
		AgentID:     agentID,
		TemplateKey: AIAgentOnboardingTemplateSchedule,
		RequestID:   "req-receipt-snapshot-1",
	}); err != nil {
		t.Fatalf("completion: %v", err)
	}
	if _, err := store.DeleteAIAgent(ctx, principal, agentID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	snapshot, err := store.snapshot(time.Now().UTC())
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	for key := range snapshot.TaskRequestReceipts {
		if strings.HasPrefix(key, onboardingCompletionReceiptKeyPrefix) {
			t.Fatalf("onboarding completion receipt %q survived agent deletion in the snapshot", key)
		}
	}
}
