package riidoaiserver

import (
	"testing"
)

func TestOnboardingAgentCreateDynamoSurvivesStaleOrdinaryWriter(t *testing.T) {
	fixture := newMembershipSharedDynamoFixture(t)
	writer, _ := newMembershipDynamoReplica(t, fixture)
	stale, _ := newMembershipDynamoReplica(t, fixture)
	ctx := withOnboardingAgentCreateRequestID(t.Context(), "stale-writer-create")
	principal := profileColorTestPrincipal()
	req := profileColorTestRequest("creation must survive stale profile writer")
	created, err := writer.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", req)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err := stale.UpdateAIAgentConfiguration(t.Context(), principal, "agent-owned-claude", UpdateAgentConfigurationRequest{Name: "unrelated profile update"}); err != nil {
			t.Fatal(err)
		}
	}
	fresh, _ := newMembershipDynamoReplica(t, fixture)
	if _, exists := fresh.agents[created.Agent.AgentID]; !exists {
		t.Fatal("ordinary stale writer erased the newly created agent")
	}
	before := len(fresh.agents)
	replayed, err := fresh.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", req)
	if err != nil || replayed.Agent.AgentID != created.Agent.AgentID || len(fresh.agents) != before {
		t.Fatalf("ordinary stale writer lost creation ownership: %v", err)
	}
}

func TestOnboardingAgentCreateDynamoDeleteDoesNotReleaseRequestKey(t *testing.T) {
	fixture := newMembershipSharedDynamoFixture(t)
	writer, _ := newMembershipDynamoReplica(t, fixture)
	principal := profileColorTestPrincipal()
	ctx := withOnboardingAgentCreateRequestID(t.Context(), "deleted-create-does-not-reappear")
	req := profileColorTestRequest("delete the created agent")
	created, err := writer.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", req)
	if err != nil {
		t.Fatal(err)
	}
	stale, _ := newMembershipDynamoReplica(t, fixture)
	if _, err := writer.DeleteAIAgent(t.Context(), principal, created.Agent.AgentID); err != nil {
		t.Fatal(err)
	}
	if _, err := stale.UpdateAIAgentConfiguration(t.Context(), principal, "agent-owned-claude", UpdateAgentConfigurationRequest{Name: "stale after delete"}); err != nil {
		t.Fatal(err)
	}
	fresh, _ := newMembershipDynamoReplica(t, fixture)
	if _, exists := fresh.agents[created.Agent.AgentID]; exists {
		t.Fatal("ordinary stale writer resurrected a deleted keyed creation")
	}
	if _, err := fresh.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", req); err == nil {
		t.Fatal("replaying the old key recreated a deliberately deleted agent")
	}
}
