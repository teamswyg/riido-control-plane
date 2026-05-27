package riidoaiserver

import (
	"context"
	"testing"
)

func TestAgentCatalogRBACAdminSeesAndMutatesPublicAndPrivateAgents(t *testing.T) {
	principal := AgentCatalogPrincipal{PrincipalID: "admin-1", Roles: []AgentCatalogRole{AgentCatalogRoleAdmin}}
	records := agentCatalogRBACRecords()

	visible := VisibleAgentCatalogRecords(principal, records)
	if got, want := agentCatalogIDs(visible), []string{"own-private", "own-public", "other-public", "other-private"}; !sameStrings(got, want) {
		t.Fatalf("admin visible agents = %v, want %v", got, want)
	}
	for _, action := range []AgentCatalogAction{AgentCatalogActionRead, AgentCatalogActionUpdate, AgentCatalogActionDelete} {
		decision := EvaluateAgentCatalogAccess(principal, records[3], action)
		if !decision.Allowed || decision.Reason != AgentCatalogDecisionAdminRole {
			t.Fatalf("admin action %s decision = %+v", action, decision)
		}
	}
}

func TestAgentCatalogRBACOwnerIsAdminForOwnAgents(t *testing.T) {
	principal := AgentCatalogPrincipal{PrincipalID: "user-1"}
	record := AgentCatalogRecord{AgentID: "own-private", OwnerPrincipalID: "user-1", Visibility: AgentCatalogVisibilityPrivate}

	for _, action := range []AgentCatalogAction{AgentCatalogActionRead, AgentCatalogActionUpdate, AgentCatalogActionDelete} {
		decision := EvaluateAgentCatalogAccess(principal, record, action)
		if !decision.Allowed || decision.Reason != AgentCatalogDecisionOwner {
			t.Fatalf("owner action %s decision = %+v", action, decision)
		}
	}
}

func TestAgentCatalogRBACNonAdminSeesOwnedAndOtherPublicAgents(t *testing.T) {
	principal := AgentCatalogPrincipal{PrincipalID: "user-1"}
	records := agentCatalogRBACRecords()

	visible := VisibleAgentCatalogRecords(principal, records)
	if got, want := agentCatalogIDs(visible), []string{"own-private", "own-public", "other-public"}; !sameStrings(got, want) {
		t.Fatalf("normal user visible agents = %v, want %v", got, want)
	}

	privateDecision := EvaluateAgentCatalogAccess(principal, records[3], AgentCatalogActionRead)
	if privateDecision.Allowed || privateDecision.Reason != AgentCatalogDecisionPrivateAgentDenied {
		t.Fatalf("other private read decision = %+v", privateDecision)
	}
}

func TestAgentCatalogRBACPublicVisibilityDoesNotGrantMutation(t *testing.T) {
	principal := AgentCatalogPrincipal{PrincipalID: "user-1"}
	publicOther := AgentCatalogRecord{AgentID: "other-public", OwnerPrincipalID: "user-2", Visibility: AgentCatalogVisibilityPublic}

	readDecision := EvaluateAgentCatalogAccess(principal, publicOther, AgentCatalogActionRead)
	if !readDecision.Allowed || readDecision.Reason != AgentCatalogDecisionPublic {
		t.Fatalf("public read decision = %+v", readDecision)
	}
	for _, action := range []AgentCatalogAction{AgentCatalogActionUpdate, AgentCatalogActionDelete} {
		decision := EvaluateAgentCatalogAccess(principal, publicOther, action)
		if decision.Allowed || decision.Reason != AgentCatalogDecisionMutationDenied {
			t.Fatalf("public mutation %s decision = %+v", action, decision)
		}
	}
}

func TestStaticTokenAuthorizerReturnsAgentCatalogRoles(t *testing.T) {
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "admin-1",
		Token:       "admin-token",
		Scopes:      []string{"riido:*"},
		Roles:       []AgentCatalogRole{AgentCatalogRoleAdmin},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	result, err := authorizer.Authorize(context.Background(), "admin-token", AuthorizationRequest{
		Resource: AuthorizationResourceMetrics,
		Action:   AuthorizationActionRead,
	})
	if err != nil {
		t.Fatalf("Authorize admin token: %v", err)
	}
	principal := AgentCatalogPrincipalFromAuthorization(result)
	if !principal.HasRole(AgentCatalogRoleAdmin) {
		t.Fatalf("principal roles = %+v", principal.Roles)
	}
}

func agentCatalogRBACRecords() []AgentCatalogRecord {
	return []AgentCatalogRecord{
		{AgentID: "own-private", OwnerPrincipalID: "user-1", Visibility: AgentCatalogVisibilityPrivate},
		{AgentID: "own-public", OwnerPrincipalID: "user-1", Visibility: AgentCatalogVisibilityPublic},
		{AgentID: "other-public", OwnerPrincipalID: "user-2", Visibility: AgentCatalogVisibilityPublic},
		{AgentID: "other-private", OwnerPrincipalID: "user-2", Visibility: AgentCatalogVisibilityPrivate},
	}
}

func agentCatalogIDs(records []AgentCatalogRecord) []string {
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.AgentID)
	}
	return ids
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
