package riidoaiserver

import "testing"

func TestAgentCatalogRolesNormalizeTrimmedAdmin(t *testing.T) {
	principal := AgentCatalogPrincipal{
		PrincipalID: " user-1 ",
		Roles:       []AgentCatalogRole{" admin "},
	}
	if err := principal.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	normalized := normalizeAgentCatalogPrincipal(principal)
	if normalized.PrincipalID != "user-1" {
		t.Fatalf("principal_id = %q", normalized.PrincipalID)
	}
	if !normalized.HasRole(AgentCatalogRoleAdmin) {
		t.Fatalf("expected trimmed admin role")
	}
}

func TestAgentCatalogRolesRejectInvalidRoleSets(t *testing.T) {
	tests := []struct {
		name  string
		roles []AgentCatalogRole
	}{
		{name: "blank role", roles: []AgentCatalogRole{""}},
		{name: "unsupported role", roles: []AgentCatalogRole{"member"}},
		{name: "duplicate admin", roles: []AgentCatalogRole{AgentCatalogRoleAdmin, " admin "}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			principal := AgentCatalogPrincipal{PrincipalID: "user-1", Roles: tt.roles}
			if err := principal.Validate(); err == nil {
				t.Fatalf("Validate succeeded for roles %v", tt.roles)
			}
			if principal.HasRole(AgentCatalogRoleAdmin) {
				t.Fatalf("HasRole returned true for invalid roles %v", tt.roles)
			}
			if _, err := normalizeAgentCatalogRoles(tt.roles); err == nil {
				t.Fatalf("normalizeAgentCatalogRoles succeeded for %v", tt.roles)
			}
		})
	}
}
