package main

import "testing"

func TestVerifyRejectsHeaderAndLoopDrift(t *testing.T) {
	for name, mutate := range map[string]func(*manifest){
		"identity": func(m *manifest) { m.ID = "other" },
		"required": func(m *manifest) { m.OwnerPackage = "" },
		"loop":     func(m *manifest) { m.Loop.Execute = "" },
	} {
		t.Run(name, func(t *testing.T) {
			m := assignmentJournalFixture()
			mutate(&m)
			if err := verify(t.TempDir(), m, false); err == nil {
				t.Fatalf("expected verify error")
			}
		})
	}
}

func TestVerifyRejectsDomainCoverageDrift(t *testing.T) {
	for name, mutate := range map[string]func(*manifest){
		"port":     func(m *manifest) { m.Ports[0].Role = "" },
		"record":   func(m *manifest) { m.Records[0].Name = "other" },
		"rule":     func(m *manifest) { m.ReplayRules[0].Rule = "" },
		"constant": func(m *manifest) { m.VersionConstants[0].Value = "" },
	} {
		t.Run(name, func(t *testing.T) {
			m := assignmentJournalFixture()
			mutate(&m)
			if err := verify(t.TempDir(), m, false); err == nil {
				t.Fatalf("expected domain drift error")
			}
		})
	}
}

func TestVerifyRejectsSourceCoverageDrift(t *testing.T) {
	for name, mutate := range map[string]func(*manifest){
		"missing": func(m *manifest) { m.SourceChecks = nil },
		"shape":   func(m *manifest) { m.SourceChecks[0].Contains = nil },
		"file":    func(m *manifest) { m.SourceChecks[0].File = "missing.go" },
		"token":   func(m *manifest) { m.SourceChecks[0].Contains = []string{"missing-token"} },
	} {
		t.Run(name, func(t *testing.T) {
			m := assignmentJournalFixture()
			repo := writeAssignmentJournalRepo(t, m)
			mutate(&m)
			if err := verify(repo, m, false); err == nil {
				t.Fatalf("expected source drift error")
			}
		})
	}
}
