package main

import "testing"

func TestGeneratedDocBindingChangeRequiresBoundFileChange(t *testing.T) {
	base := []claimBinding{{
		ID: "claim", Statement: "same", Loop: "loop",
		Files: []string{"internal/example.go"}, GeneratedDoc: []string{"docs/old.md"},
	}}
	current := []claimBinding{{
		ID: "claim", Statement: "same", Loop: "loop",
		Files: []string{"internal/example.go"}, GeneratedDoc: []string{"docs/new.md"},
	}}
	if _, err := verifyClaimImpact("origin/main", base, current,
		map[string]bool{defaultManifest: true}); err == nil {
		t.Fatal("expected generated doc binding change without bound file to fail")
	}
}
