package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOnboardingAgentCreateReplayWorksWithoutRuntimeAndRejectsDeletedAgent(t *testing.T) {
	ctx := withOnboardingAgentCreateRequestID(t.Context(), "runtime-disappeared")
	principal := profileColorTestPrincipal()
	store := NewDevelopmentAIAgentClientStore()
	req := profileColorTestRequest("runtime disappears after creation")
	first, err := store.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", req)
	if err != nil {
		t.Fatal(err)
	}
	store.devices = nil
	replayed, err := store.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", req)
	if err != nil || replayed.Agent.AgentID != first.Agent.AgentID {
		t.Fatalf("replay revalidated a disappeared runtime instead of returning durable creation: %v", err)
	}
	if _, err := store.DeleteAIAgent(t.Context(), principal, first.Agent.AgentID); err != nil {
		t.Fatal(err)
	}
	_, err = store.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", req)
	if err == nil {
		t.Fatal("deleted creation replay succeeded")
	}
	rec := httptest.NewRecorder()
	writeAIAgentClientError(rec, err)
	if rec.Code != http.StatusConflict {
		t.Fatalf("deleted creation replay error=%v status=%d, want 409 without resurrection", err, rec.Code)
	}
}
