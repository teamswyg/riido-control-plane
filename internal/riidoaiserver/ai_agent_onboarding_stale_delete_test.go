package riidoaiserver

import (
	"context"
	"testing"
)

func TestOnboardingDeleteReloadsCompletionUnknownToItsReplica(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	fixture := newMembershipSharedDynamoFixture(t)
	writer, _ := newMembershipDynamoReplica(t, fixture)
	created, err := writer.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("old incarnation"))
	if err != nil {
		t.Fatal(err)
	}
	stale, _ := newMembershipDynamoReplica(t, fixture)
	request := CompleteAIAgentOnboardingTaskRequest{AgentID: created.Agent.AgentID, TemplateKey: AIAgentOnboardingTemplateSchedule, RequestID: "deleted-incarnation-receipt"}
	if _, err := writer.CompleteAIAgentOnboardingTask(ctx, principal, "task-old-incarnation", request); err != nil {
		t.Fatal(err)
	}
	if _, err := stale.DeleteAIAgent(ctx, principal, created.Agent.AgentID); err != nil {
		t.Fatal(err)
	}
	recreated, err := stale.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("new incarnation"))
	if err != nil {
		t.Fatal(err)
	}
	if recreated.Agent.AgentID != created.Agent.AgentID {
		t.Fatal("fixture did not exercise same-id reuse")
	}
	if _, err := stale.CompleteAIAgentOnboardingTask(ctx, principal, "task-new-incarnation", request); err != nil {
		t.Fatalf("deleted incarnation completion or receipt was resurrected: %v", err)
	}
}
