package main

import (
	"reflect"
	"testing"
)

func TestClaimImpactEvidenceExposesChangedFiles(t *testing.T) {
	current := []claimBinding{{
		ID:           "claim",
		Statement:    "kept",
		Loop:         "loop",
		Files:        []string{"internal/example.go"},
		GeneratedDoc: []string{"docs/claim.md"},
	}}
	evidence, err := verifyClaimImpact("origin/main", current, current, map[string]bool{
		"internal/example.go": true,
		"docs/claim.md":       true,
	})
	if err != nil {
		t.Fatalf("verify impact: %v", err)
	}
	want := []string{"docs/claim.md", "internal/example.go"}
	if !reflect.DeepEqual(evidence.ChangedFiles, want) {
		t.Fatalf("changed files = %#v, want %#v", evidence.ChangedFiles, want)
	}
}
