package main

import "testing"

func TestVerifierCommandsDerivePackageCommands(t *testing.T) {
	got := verifierCommands([]ref{
		{Kind: "test", Path: "tools/evidencegraph/impact_test.go"},
		{Kind: "tool", Path: "tools/loopregistry"},
		{Kind: "workflow", Path: ".github/workflows/evidence-graph.yml"},
		{Kind: "test", Path: "tools/evidencegraph/impact_test.go"},
	})
	want := []string{
		"go test ./tools/evidencegraph -count=1",
		"go test ./tools/loopregistry -count=1",
	}
	if len(got) != len(want) {
		t.Fatalf("commands = %+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("commands = %+v", got)
		}
	}
}
