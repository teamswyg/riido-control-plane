package main

import "testing"

func TestVerifyRejectsHeaderDomainProfileAndLoopDrift(t *testing.T) {
	for name, mutate := range map[string]func(*manifest){
		"schema":     func(m *manifest) { m.SchemaVersion = "other" },
		"required":   func(m *manifest) { m.GeneratedDoc = "" },
		"role":       func(m *manifest) { m.Roles = nil },
		"visibility": func(m *manifest) { m.Visibilities = []string{"public"} },
		"action":     func(m *manifest) { m.Actions = []string{"read"} },
		"rule":       func(m *manifest) { m.VisibilityRules[0].ID = "other" },
		"route":      func(m *manifest) { m.Routes = nil },
		"profile":    func(m *manifest) { m.EvidenceProfiles = nil },
		"profile-shape": func(m *manifest) {
			m.EvidenceProfiles[0].Focus = ""
		},
		"loop": func(m *manifest) { m.Loop.Evaluate = "" },
	} {
		t.Run(name, func(t *testing.T) {
			m := agentCatalogRBACFixture()
			mutate(&m)
			if err := verify(t.TempDir(), m, false); err == nil {
				t.Fatalf("expected verification error")
			}
		})
	}
}

func TestVerifyRejectsProfileSourceAndDocDrift(t *testing.T) {
	for name, mutate := range map[string]func(*manifest){
		"profile-file":     func(m *manifest) { m.EvidenceProfiles[0].Workflow = "missing.yml" },
		"profile-artifact": func(m *manifest) { m.EvidenceProfiles[0].EvidenceArtifact = "missing" },
		"profile-pattern":  func(m *manifest) { m.EvidenceProfiles[0].TestPatterns = []string{"missing"} },
		"source-checks":    func(m *manifest) { m.SourceChecks = nil },
		"source-shape":     func(m *manifest) { m.SourceChecks[0].Contains = nil },
		"source-file":      func(m *manifest) { m.SourceChecks[0].File = "missing.go" },
		"source-token":     func(m *manifest) { m.SourceChecks[0].Contains = []string{"missing"} },
	} {
		t.Run(name, func(t *testing.T) {
			m := agentCatalogRBACFixture()
			repo := writeAgentCatalogRepo(t, m)
			mutate(&m)
			if err := verify(repo, m, false); err == nil {
				t.Fatalf("expected source or profile error")
			}
		})
	}
	m := agentCatalogRBACFixture()
	repo := writeAgentCatalogRepo(t, m)
	if err := verify(repo, m, true); err == nil {
		t.Fatalf("expected missing generated doc error")
	}
	writeAgentCatalogFile(t, repo+"/"+m.GeneratedDoc, "stale")
	if err := verify(repo, m, true); err == nil {
		t.Fatalf("expected generated doc drift")
	}
}
