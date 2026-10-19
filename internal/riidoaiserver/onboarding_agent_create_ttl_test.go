package riidoaiserver

import (
	"testing"
	"time"
)

func TestOnboardingAgentCreateExpiredReceiptDoesNotBlockNewIntent(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	ctx := withOnboardingAgentCreateRequestID(t.Context(), "expired-create-request")
	first, err := store.CreateAIAgentFromOnboardingFixture(ctx, profileColorTestPrincipal(), "riido_pm", profileColorTestRequest("old draft"))
	if err != nil {
		t.Fatal(err)
	}
	for key, receipt := range store.onboardingAgentCreates {
		receipt.CreatedAt = time.Now().UTC().Add(-25 * time.Hour)
		store.onboardingAgentCreates[key] = receipt
	}
	second, err := store.CreateAIAgentFromOnboardingFixture(ctx, profileColorTestPrincipal(), "riido_pm", profileColorTestRequest("new draft after documented retention"))
	if err != nil || second.Agent.AgentID == first.Agent.AgentID || len(store.onboardingAgentCreates) != 1 {
		t.Fatalf("expired receipt prevented a new creation: %v", err)
	}
}

func TestOnboardingAgentCreateDynamoOrdinaryWriterDoesNotRestoreExpiredReceipts(t *testing.T) {
	fixture := newMembershipSharedDynamoFixture(t)
	writer, _ := newMembershipDynamoReplica(t, fixture)
	principal := profileColorTestPrincipal()
	ctx := withOnboardingAgentCreateRequestID(t.Context(), "old-key-pruned-on-next-create")
	first, err := writer.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("old retained creation"))
	if err != nil {
		t.Fatal(err)
	}
	for key, receipt := range writer.onboardingAgentCreates {
		receipt.CreatedAt = time.Now().UTC().Add(-25 * time.Hour)
		writer.onboardingAgentCreates[key] = receipt
	}
	if err := writer.saveSnapshot(t.Context()); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.CreateAIAgentFromOnboardingFixture(withOnboardingAgentCreateRequestID(t.Context(), "fresh-key-after-prune"), principal, "riido_pm", profileColorTestRequest("fresh retained creation")); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.UpdateAIAgentConfiguration(t.Context(), principal, "agent-owned-claude", UpdateAgentConfigurationRequest{Name: "ordinary write after receipt pruning"}); err != nil {
		t.Fatal(err)
	}
	fresh, _ := newMembershipDynamoReplica(t, fixture)
	for _, receipt := range fresh.onboardingAgentCreates {
		if receipt.AgentID == first.Agent.AgentID {
			t.Fatal("ordinary writer restored a receipt already removed by retention")
		}
	}
}
