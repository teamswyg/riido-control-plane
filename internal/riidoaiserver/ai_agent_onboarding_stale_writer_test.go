package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

// A profile update from a replica that predates completion must not erase the
// durable one-time completion record and permit another completion after reload.
func TestOnboardingCompletionSurvivesStaleOrdinarySnapshotWriter(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	fixture := newMembershipSharedDynamoFixture(t)
	writer, _ := newMembershipDynamoReplica(t, fixture)
	created, err := writer.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("one-time completion"))
	if err != nil {
		t.Fatal(err)
	}
	other, err := writer.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("second completion target"))
	if err != nil {
		t.Fatal(err)
	}
	stale, _ := newMembershipDynamoReplica(t, fixture)
	request := CompleteAIAgentOnboardingTaskRequest{AgentID: created.Agent.AgentID, TemplateKey: AIAgentOnboardingTemplateSchedule, RequestID: "one-time-original"}
	first, err := writer.CompleteAIAgentOnboardingTask(ctx, principal, "task-completed-once", request)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err := stale.UpdateAIAgentConfiguration(ctx, principal, "agent-owned-claude", UpdateAgentConfigurationRequest{Name: "Unrelated profile update"}); err != nil {
			t.Fatal(err)
		}
	}
	fresh, _ := newMembershipDynamoReplica(t, fixture)
	if threads := fresh.taskThreads["task-completed-once"]; len(threads) != 1 || threads[0].AssignmentID != first.AssignmentID {
		t.Fatal("stale ordinary writer erased the visible completion comment")
	}
	if _, err := fresh.CompleteAIAgentOnboardingTask(ctx, principal, "task-must-not-complete-twice", CompleteAIAgentOnboardingTaskRequest{AgentID: created.Agent.AgentID, TemplateKey: AIAgentOnboardingTemplateSchedule}); !errors.Is(err, errOnboardingCompletionConflict) {
		t.Fatalf("stale ordinary writer lost the one-time completion fence: %v", err)
	}
	replayed, err := fresh.CompleteAIAgentOnboardingTask(ctx, principal, "task-completed-once", request)
	if err != nil {
		t.Fatal(err)
	}
	assertSameJSON(t, first, replayed)
	if _, err := fresh.CompleteAIAgentOnboardingTask(ctx, principal, "task-receipt-must-not-be-reused", CompleteAIAgentOnboardingTaskRequest{AgentID: other.Agent.AgentID, TemplateKey: AIAgentOnboardingTemplateSchedule, RequestID: request.RequestID}); !errors.Is(err, errOnboardingCompletionConflict) {
		t.Fatalf("stale ordinary writer erased request-id ownership: %v", err)
	}
}
