package riidoaiserver

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestOnboardingDeleteClearsCompletionCommittedAfterItsSnapshotLoad(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	fixture := newMembershipSharedDynamoFixture(t)
	writer, _ := newMembershipDynamoReplica(t, fixture)
	created, err := writer.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("delete overlapping completion"))
	if err != nil {
		t.Fatal(err)
	}
	deleter, _ := newMembershipDynamoReplica(t, fixture)
	entered, resume := make(chan struct{}), make(chan struct{})
	var armed atomic.Bool
	armed.Store(true)
	fixture.mu.Lock()
	fixture.beforeOwnedSnapshotRead = func() {
		if armed.Swap(false) {
			close(entered)
			<-resume
		}
	}
	fixture.mu.Unlock()
	deleted := make(chan error, 1)
	go func() { _, err := deleter.DeleteAIAgent(ctx, principal, created.Agent.AgentID); deleted <- err }()
	<-entered
	request := CompleteAIAgentOnboardingTaskRequest{AgentID: created.Agent.AgentID, TemplateKey: AIAgentOnboardingTemplateSchedule, RequestID: "overlapping-deleted-receipt"}
	_, completionErr := writer.CompleteAIAgentOnboardingTask(ctx, principal, "task-old-overlap", request)
	close(resume)
	if completionErr != nil {
		t.Fatal(completionErr)
	}
	if err := <-deleted; err != nil {
		t.Fatal(err)
	}
	fresh, _ := newMembershipDynamoReplica(t, fixture)
	if _, err := fresh.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("new incarnation after overlap")); err != nil {
		t.Fatal(err)
	}
	if _, err := fresh.CompleteAIAgentOnboardingTask(ctx, principal, "task-new-overlap", request); err != nil {
		t.Fatalf("deletion preserved an overlapping old completion or receipt: %v", err)
	}
}
