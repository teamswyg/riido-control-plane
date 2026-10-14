package main

import "testing"

func TestVerifyRejectsSourceCaseAndDocDrift(t *testing.T) {
	for name, mutate := range map[string]func(*manifest, []corsEvidence){
		"source-file":  func(m *manifest, _ []corsEvidence) { m.SourceChecks[0].File = "missing.go" },
		"source-token": func(m *manifest, _ []corsEvidence) { m.SourceChecks[0].Contains = []string{"missing"} },
		"case-count":   func(_ *manifest, r []corsEvidence) { r[0].Name = "" },
		"case-name":    func(m *manifest, _ []corsEvidence) { m.CORSCases[0].Name = "missing" },
	} {
		t.Run(name, func(t *testing.T) {
			m := webFrontendAPIFixture()
			repo := writeWebFrontendRepo(t, m)
			results, err := verifyCORSCases(m.CORSCases)
			if err != nil {
				t.Fatal(err)
			}
			mutate(&m, results)
			if err := verify(repo, m, results, false); err == nil {
				t.Fatalf("expected verification error")
			}
		})
	}
	m := webFrontendAPIFixture()
	repo := writeWebFrontendRepo(t, m)
	results, err := verifyCORSCases(m.CORSCases)
	if err != nil {
		t.Fatal(err)
	}
	if err := verify(repo, m, results, true); err == nil {
		t.Fatalf("expected missing generated doc error")
	}
	writeWebFrontendFile(t, repo+"/"+m.GeneratedDoc, "stale")
	if err := verify(repo, m, results, true); err == nil {
		t.Fatalf("expected generated doc drift")
	}
}
