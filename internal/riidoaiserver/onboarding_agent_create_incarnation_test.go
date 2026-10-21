package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOnboardingAgentCreateReplayRejectsReusedAgentIDIncarnation(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := profileColorTestPrincipal()
	ctx := withOnboardingAgentCreateRequestID(t.Context(), "old-creation-incarnation")
	req := profileColorTestRequest("original keyed result")
	original, err := store.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", req)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.DeleteAIAgent(t.Context(), principal, original.Agent.AgentID); err != nil {
		t.Fatal(err)
	}
	replacement, err := store.CreateAIAgentFromOnboardingFixture(t.Context(), principal, "riido_pm", profileColorTestRequest("new independent incarnation"))
	if err != nil {
		t.Fatal(err)
	}
	if replacement.Agent.AgentID != original.Agent.AgentID {
		t.Fatal("test did not exercise the reused-id boundary")
	}
	if replacement.Agent.CreatedAt.Equal(original.Agent.CreatedAt) {
		t.Fatal("test did not create a distinct agent incarnation")
	}
	_, err = store.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", req)
	if err == nil {
		t.Fatal("old request replay returned a different incarnation sharing its agent id")
	}
	rec := httptest.NewRecorder()
	writeAIAgentClientError(rec, err)
	if rec.Code != http.StatusConflict || store.agents[replacement.Agent.AgentID].Name != "new independent incarnation" {
		t.Fatal("old request replay changed the replacement agent or lost its conflict")
	}
}
