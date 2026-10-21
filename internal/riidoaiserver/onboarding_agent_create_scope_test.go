package riidoaiserver

import (
	"testing"
)

func TestOnboardingAgentCreateRequestKeyIsolatedByPrincipalAndWorkspace(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	ctx := withOnboardingAgentCreateRequestID(t.Context(), "same-key-different-owner")
	principals := []AuthorizationResult{
		profileColorTestPrincipal(),
		{PrincipalID: "user-2", WorkspaceID: defaultAIAgentClientWorkspaceID, Roles: []AgentCatalogRole{AgentCatalogRoleAdmin}},
		{PrincipalID: "user-1", WorkspaceID: "different-workspace"},
	}
	ids := map[string]bool{}
	for _, principal := range principals {
		response, err := store.CreateAIAgentFromOnboardingFixture(ctx, principal, "riido_pm", profileColorTestRequest("same display name"))
		if err != nil {
			t.Fatal(err)
		}
		if ids[response.Agent.AgentID] || response.Agent.OwnerPrincipalID != principal.PrincipalID || response.Agent.WorkspaceID != principal.WorkspaceID {
			t.Fatal("a request key crossed a principal or workspace boundary")
		}
		ids[response.Agent.AgentID] = true
	}
}
