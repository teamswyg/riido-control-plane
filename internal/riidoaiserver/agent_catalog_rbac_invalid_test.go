package riidoaiserver

import "testing"

func TestAgentCatalogRBACDeniesInvalidInputs(t *testing.T) {
	validPrincipal := AgentCatalogPrincipal{PrincipalID: "user-1"}
	validRecord := AgentCatalogRecord{
		AgentID:          "agent-1",
		OwnerPrincipalID: "user-1",
		Visibility:       AgentCatalogVisibilityPrivate,
	}
	tests := []struct {
		name      string
		principal AgentCatalogPrincipal
		record    AgentCatalogRecord
		action    AgentCatalogAction
		reason    AgentCatalogAccessDecisionReason
	}{
		{
			name:      "blank principal",
			principal: AgentCatalogPrincipal{},
			record:    validRecord,
			action:    AgentCatalogActionRead,
			reason:    AgentCatalogDecisionInvalidPrincipal,
		},
		{
			name:      "blank agent",
			principal: validPrincipal,
			record:    AgentCatalogRecord{OwnerPrincipalID: "user-1", Visibility: AgentCatalogVisibilityPrivate},
			action:    AgentCatalogActionRead,
			reason:    AgentCatalogDecisionInvalidAgentRecord,
		},
		{
			name:      "blank owner",
			principal: validPrincipal,
			record:    AgentCatalogRecord{AgentID: "agent-1", Visibility: AgentCatalogVisibilityPrivate},
			action:    AgentCatalogActionRead,
			reason:    AgentCatalogDecisionInvalidAgentRecord,
		},
		{
			name:      "bad visibility",
			principal: validPrincipal,
			record:    AgentCatalogRecord{AgentID: "agent-1", OwnerPrincipalID: "user-1", Visibility: "team"},
			action:    AgentCatalogActionRead,
			reason:    AgentCatalogDecisionInvalidAgentRecord,
		},
		{
			name:      "bad action",
			principal: validPrincipal,
			record:    validRecord,
			action:    "publish",
			reason:    AgentCatalogDecisionInvalidAction,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := EvaluateAgentCatalogAccess(tt.principal, tt.record, tt.action)
			if decision.Allowed || decision.Reason != tt.reason {
				t.Fatalf("decision = %+v, want denied %s", decision, tt.reason)
			}
		})
	}
}
