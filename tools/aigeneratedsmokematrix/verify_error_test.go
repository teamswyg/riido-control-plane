package main

import "testing"

func TestVerifyManifestShapeRejectsDrift(t *testing.T) {
	for name, mutate := range map[string]func(*manifest){
		"identity": func(m *manifest) { m.SchemaVersion = "other" },
		"required": func(m *manifest) { m.GeneratedDoc = "" },
	} {
		t.Run(name, func(t *testing.T) {
			m := smokeFixtureManifest()
			mutate(&m)
			if err := verifyManifestShape(m); err == nil {
				t.Fatalf("expected manifest shape error")
			}
		})
	}
}

func TestVerifyAllRejectsBrokenInputs(t *testing.T) {
	for name, mutate := range map[string]func(*testing.T, string, *manifest){
		"counts": func(_ *testing.T, _ string, m *manifest) {
			m.OperationCounts.Total = 3
		},
		"loop": func(_ *testing.T, _ string, m *manifest) {
			m.Loop.Evaluate = ""
		},
		"source": func(_ *testing.T, _ string, m *manifest) {
			m.SourceChecks[0].Contains = []string{"missing"}
		},
		"openapi": func(_ *testing.T, _ string, m *manifest) {
			m.OpenAPI = "missing-openapi.json"
		},
		"smoke-read": func(_ *testing.T, _ string, m *manifest) {
			m.SmokeMatrix = "missing-smoke.json"
		},
		"smoke": func(t *testing.T, repo string, m *manifest) {
			t.Helper()
			mustJSON(t, repo+"/"+m.SmokeMatrix, smokeMatrix{SchemaVersion: "other"})
		},
	} {
		t.Run(name, func(t *testing.T) {
			repo, m := writeSmokeFixtureRepo(t)
			mutate(t, repo, &m)
			if _, err := verifyAll(repo, m); err == nil {
				t.Fatalf("expected verifyAll error")
			}
		})
	}
}

func TestLoadOpenAPIGeneratedRejectsInvalidJSON(t *testing.T) {
	repo, m := writeSmokeFixtureRepo(t)
	if err := writeText(repo+"/"+m.OpenAPI, "{"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadOpenAPIGenerated(repo + "/" + m.OpenAPI); err == nil {
		t.Fatalf("expected invalid OpenAPI JSON error")
	}
}
