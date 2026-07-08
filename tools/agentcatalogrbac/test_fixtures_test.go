package main

import (
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/agentcatalogrbac/requirements"
)

func agentCatalogRBACFixture() manifest {
	return manifest{
		SchemaVersion:    requirements.ManifestSchema,
		ID:               requirements.ExpectedID,
		Title:            "Agent Catalog RBAC",
		RiidoTask:        requirements.ExpectedTask,
		GeneratedDoc:     "docs/agent-catalog-rbac.md",
		Workflow:         ".github/workflows/agent-catalog-rbac.yml",
		EvidenceArtifact: "agent-catalog-rbac-evidence",
		OwnerPackage:     "internal/riidoaiserver",
		EvidenceProfiles: []profile{{
			ID: "rbac", Workflow: ".github/workflows/agent-catalog-rbac.yml",
			EvidenceArtifact: "agent-catalog-rbac-evidence", Focus: "RBAC",
			TestPatterns: []string{"AgentCatalogRBAC|StaticTokenAuthorizer"},
		}},
		Roles:               []string{"admin"},
		Visibilities:        []string{"public", "private"},
		Actions:             []string{"read", "update", "delete"},
		VisibilityRules:     agentCatalogRulesFixture(),
		AuthorizationScopes: []string{"agent-catalog:read"},
		Routes:              append([]string{}, requirements.RequiredRoutes...),
		RequestDTOs:         []string{"CreateAgentCatalogRequest"},
		ResponseDTOs:        []string{"AgentCatalogListResponse"},
		StoreMethods:        []string{"ListAgentCatalog"},
		SourceChecks:        []sourceCheck{{Name: "rbac", File: "internal/rbac.go", Contains: []string{"EvaluateAgentCatalogAccess"}}},
		Loop:                evidenceLoop{"observe", "hypothesis", "execute", "evaluate", "retro"},
	}
}

func agentCatalogRulesFixture() []rule {
	ids := []string{"admin-all", "owner-all", "public-read", "private-deny", "public-mutation-deny"}
	rules := make([]rule, 0, len(ids))
	for _, id := range ids {
		rules = append(rules, rule{ID: id, Subject: "subject", Record: "record", Reason: "reason"})
	}
	return rules
}

func writeAgentCatalogRepo(t *testing.T, m manifest) string {
	t.Helper()
	repo := t.TempDir()
	writeAgentCatalogFile(t, filepath.Join(repo, "go.mod"), "module example.com/agent\n")
	writeAgentCatalogFile(t, filepath.Join(repo, "internal/rbac.go"),
		"package internal\n// EvaluateAgentCatalogAccess\n")
	writeAgentCatalogFile(t, filepath.Join(repo, ".github/workflows/agent-catalog-rbac.yml"),
		"agent-catalog-rbac-evidence AgentCatalogRBAC|StaticTokenAuthorizer")
	writeAgentCatalogJSON(t, filepath.Join(repo, "manifest.json"), m)
	return repo
}
