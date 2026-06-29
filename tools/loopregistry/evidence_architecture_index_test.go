package main

import "testing"

func TestLoopRegistryEvidenceExposesArchitectureIndex(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	result, err := verifyAll("../..", m, hashes)
	if err != nil {
		t.Fatal(err)
	}
	got := newEvidence(m, result, nil).ArchitectureIndex
	if got.ComponentCount == 0 || got.PathCount == 0 ||
		got.BindingCount == 0 || got.VerifierCommandCount == 0 {
		t.Fatalf("empty architecture index: %+v", got)
	}
	binding := architectureBindingByPath(got.Paths, "tools/loopregistry/evidence.go")
	if binding.Path == "" {
		t.Fatal("tools/loopregistry/evidence.go missing from architecture index")
	}
	if len(binding.ClaimIDs) == 0 || len(binding.LoopIDs) == 0 ||
		len(binding.VerifierCommands) == 0 || len(binding.EvidenceChainIDs) == 0 {
		t.Fatalf("incomplete architecture binding: %+v", binding)
	}
}

func TestArchitectureIndexExposesComponentSummary(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	result, err := verifyAll("../..", m, hashes)
	if err != nil {
		t.Fatal(err)
	}
	got := newEvidence(m, result, nil).ArchitectureIndex
	component := architectureComponentByName(got.Components, "tools/loopregistry")
	if component.Component == "" {
		t.Fatal("tools/loopregistry component missing from architecture index")
	}
	if component.PathCount == 0 || len(component.ClaimIDs) == 0 ||
		len(component.VerifierCommands) == 0 || len(component.EvidenceChainIDs) == 0 {
		t.Fatalf("incomplete architecture component: %+v", component)
	}
}

func TestArchitectureIndexIncludesGeneratedDocs(t *testing.T) {
	surface := claimSurface{
		ID:               "claim-a",
		GeneratedDocs:    []string{"docs/generated.md"},
		VerifierCommands: []string{"go test ./tools/example -count=1"},
		EvidenceChainIDs: []string{"chain-a"},
	}
	index := architectureIndexFor(
		[]claimBinding{{ID: "claim-a", Loop: "closed_loop_candidate"}},
		[]claimSurface{surface},
	)
	binding := architectureBindingByPath(index.Paths, "docs/generated.md")
	if binding.Kind != "generated_doc" || binding.LoopIDs[0] != "closed_loop_candidate" {
		t.Fatalf("generated doc binding = %+v", binding)
	}
}

func TestArchitectureComponentIDUsesStablePathGrammar(t *testing.T) {
	cases := map[string]string{
		".github/workflows/loop-registry.yml": ".github/workflows",
		"tools/loopregistry/evidence.go":      "tools/loopregistry",
		"internal/riidoaiserver/server.go":    "internal/riidoaiserver",
		"docs/30-architecture/loop.md":        "docs/30-architecture",
		"go.mod":                              "go.mod",
	}
	for path, want := range cases {
		if got := architectureComponentID(path); got != want {
			t.Fatalf("architectureComponentID(%q) = %q, want %q", path, got, want)
		}
	}
}
