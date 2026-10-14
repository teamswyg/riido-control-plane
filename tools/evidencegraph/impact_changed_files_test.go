package main

import (
	"reflect"
	"testing"
)

func TestEvidenceGraphImpactExposesChangedFiles(t *testing.T) {
	current := []chain{testChain("chain", "same")}
	evidence, err := verifyChainImpact("origin/main", current, current,
		map[string]bool{
			defaultManifest:               true,
			"tools/evidencegraph/run.go":  true,
			"docs/30-architecture/doc.md": true,
		})
	if err != nil {
		t.Fatalf("verify impact: %v", err)
	}
	want := []string{
		"docs/30-architecture/doc.md",
		"docs/30-architecture/evidence-graph.riido.json",
		"tools/evidencegraph/run.go",
	}
	if !reflect.DeepEqual(evidence.ChangedFiles, want) {
		t.Fatalf("changed files = %#v, want %#v", evidence.ChangedFiles, want)
	}
}
