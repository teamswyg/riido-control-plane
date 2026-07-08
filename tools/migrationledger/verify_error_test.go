package main

import "testing"

func TestVerifyRejectsManifestDrift(t *testing.T) {
	for name, mutate := range map[string]func(*manifest){
		"identity": func(m *manifest) { m.ID = "other" },
		"required": func(m *manifest) { m.GeneratedDoc = "" },
		"loop":     func(m *manifest) { m.Loop.Evaluate = "" },
		"intro":    func(m *manifest) { m.Intro = nil },
		"assert":   func(m *manifest) { m.Assertions = nil },
	} {
		t.Run(name, func(t *testing.T) {
			m := migrationLedgerFixture()
			mutate(&m)
			if _, err := verify(m); err == nil {
				t.Fatalf("expected verify error")
			}
		})
	}
}

func TestVerifyRejectsSectionCoverageDrift(t *testing.T) {
	for name, mutate := range map[string]func(*manifest){
		"missing-required": func(m *manifest) { m.Sections[0].Title = "Other" },
		"bad-level":        func(m *manifest) { m.Sections[0].Level = 1 },
		"bad-shape":        func(m *manifest) { m.Sections[0].Kind = "" },
		"empty-slice":      func(m *manifest) { m.Sections[8].Body = nil },
		"too-few-slices":   func(m *manifest) { m.Sections = m.Sections[:20] },
		"too-few-gates":    func(m *manifest) { m.Sections[5].Body = []string{"- gate"} },
		"too-few-refs":     func(m *manifest) { clearMigrationRefs(m) },
	} {
		t.Run(name, func(t *testing.T) {
			m := migrationLedgerFixture()
			mutate(&m)
			if _, err := verify(m); err == nil {
				t.Fatalf("expected section coverage error")
			}
		})
	}
}

func clearMigrationRefs(m *manifest) {
	for i := range m.Sections {
		m.Sections[i].Body = []string{"body"}
		if m.Sections[i].Kind == "slice" {
			m.Sections[i].Title = "slice"
		}
	}
}
