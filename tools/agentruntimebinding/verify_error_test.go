package main

import "testing"

func TestVerifyRejectsHeaderDomainAndLoopDrift(t *testing.T) {
	for name, mutate := range map[string]func(*manifest){
		"schema":       func(m *manifest) { m.SchemaVersion = "other" },
		"required":     func(m *manifest) { m.GeneratedDoc = "" },
		"field":        func(m *manifest) { m.BindingFields[0].Required = false },
		"device-field": func(m *manifest) { m.BindingFields[2].Required = true },
		"binding-rule": func(m *manifest) { m.BindingRules[0].Rule = "" },
		"device-rule":  func(m *manifest) { m.DeviceRules[0].Rule = "" },
		"loop":         func(m *manifest) { m.Loop.Evaluate = "" },
	} {
		t.Run(name, func(t *testing.T) {
			m := agentRuntimeBindingFixture()
			mutate(&m)
			if err := verify(t.TempDir(), m, false); err == nil {
				t.Fatalf("expected verification error")
			}
		})
	}
}

func TestVerifyRejectsSourceAndDocDrift(t *testing.T) {
	for name, mutate := range map[string]func(*manifest, string){
		"missing-checks": func(m *manifest, _ string) { m.SourceChecks = nil },
		"check-shape":    func(m *manifest, _ string) { m.SourceChecks[0].Contains = nil },
		"check-file":     func(m *manifest, _ string) { m.SourceChecks[0].File = "missing.go" },
		"check-token":    func(m *manifest, _ string) { m.SourceChecks[0].Contains = []string{"missing"} },
	} {
		t.Run(name, func(t *testing.T) {
			m := agentRuntimeBindingFixture()
			repo := writeAgentRuntimeBindingRepo(t, m)
			mutate(&m, repo)
			if err := verify(repo, m, false); err == nil {
				t.Fatalf("expected source error")
			}
		})
	}
	m := agentRuntimeBindingFixture()
	repo := writeAgentRuntimeBindingRepo(t, m)
	if err := verify(repo, m, true); err == nil {
		t.Fatalf("expected missing generated doc error")
	}
	writeAgentRuntimeBindingFile(t, repo+"/"+m.GeneratedDoc, "stale")
	if err := verify(repo, m, true); err == nil {
		t.Fatalf("expected generated doc drift")
	}
}
