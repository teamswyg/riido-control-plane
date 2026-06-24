package main

import "testing"

func TestClaimImpactRequiresBoundFileChange(t *testing.T) {
	base := []claimBinding{{
		ID:        "claim",
		Statement: "old",
		Loop:      "loop",
		Files:     []string{"internal/example.go"},
	}}
	current := []claimBinding{{
		ID:        "claim",
		Statement: "new",
		Loop:      "loop",
		Files:     []string{"internal/example.go"},
	}}
	if _, err := verifyClaimImpact("origin/main", base, current, map[string]bool{}); err == nil {
		t.Fatal("expected missing bound file change to fail")
	}
}

func TestClaimImpactAllowsBoundFileChange(t *testing.T) {
	base := []claimBinding{{ID: "claim", Statement: "old", Loop: "loop", Files: []string{"a.go"}}}
	current := []claimBinding{{ID: "claim", Statement: "new", Loop: "loop", Files: []string{"a.go"}}}
	evidence, err := verifyClaimImpact("origin/main", base, current, map[string]bool{
		"a.go": true, defaultManifest: true,
	})
	if err != nil {
		t.Fatalf("verify impact: %v", err)
	}
	if evidence.ChangedClaimCount != 1 || len(evidence.Claims[0].ChangedBoundFiles) != 1 {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
}

func TestBoundFileImpactRequiresClaimEvidenceChange(t *testing.T) {
	current := []claimBinding{{
		ID: "claim", Statement: "kept", Loop: "loop",
		Files: []string{"internal/example.go"}, GeneratedDoc: []string{"docs/claim.md"},
	}}
	if _, err := verifyClaimImpact("origin/main", current, current,
		map[string]bool{"internal/example.go": true}); err == nil {
		t.Fatal("expected bound file change without evidence to fail")
	}
}

func TestBoundFileImpactAllowsGeneratedDocChange(t *testing.T) {
	current := []claimBinding{{
		ID: "claim", Statement: "kept", Loop: "loop",
		Files: []string{"internal/example.go"}, GeneratedDoc: []string{"docs/claim.md"},
	}}
	evidence, err := verifyClaimImpact("origin/main", current, current,
		map[string]bool{"internal/example.go": true, "docs/claim.md": true})
	if err != nil {
		t.Fatalf("verify impact: %v", err)
	}
	if evidence.BoundSurfaceChangeCount != 1 || len(evidence.BoundSurfaces[0].ChangedEvidence) != 1 {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
}
