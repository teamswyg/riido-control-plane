package main

import "testing"

func TestVerifyAllRejectsManifestIdentityAndShape(t *testing.T) {
	t.Parallel()
	for name, mutate := range map[string]func(*manifest){
		"identity": func(m *manifest) { m.SchemaVersion = "wrong" },
		"binding":  func(m *manifest) { m.GeneratedDoc = "" },
		"surface":  func(m *manifest) { m.Surfaces = nil },
		"category": func(m *manifest) { m.RequiredCategories = nil },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			m := validManifestForVerifyTest()
			mutate(&m)
			if err := verifyAll("../..", m); err == nil {
				t.Fatalf("verifyAll accepted invalid %s manifest", name)
			}
		})
	}
}

func validManifestForVerifyTest() manifest {
	m := loadFixtureManifest()
	m.GeneratedDoc = "docs/30-architecture/control-plane-high-traffic-audit.md"
	m.Workflow = ".github/workflows/control-plane-performance.yml"
	m.Surfaces = []surface{{
		ID:        "server",
		Category:  "endpoint_hot_path",
		Risk:      "risk",
		Files:     []string{"internal/riidoaiserver/server.go"},
		Patterns:  []string{"http"},
		Candidate: "candidate",
	}}
	return m
}

func loadFixtureManifest() manifest {
	var m manifest
	if err := readJSON("../../"+defaultManifest, &m); err != nil {
		panic(err)
	}
	return m
}
