package main

import "testing"

func TestBoundFileImpactRequiresClaimEvidenceChange(t *testing.T) {
	current := boundSurfaceImpactClaim()
	if _, err := verifyClaimImpact("origin/main", current, current,
		map[string]bool{"internal/example.go": true}); err == nil {
		t.Fatal("expected bound file change without evidence to fail")
	}
}

func TestBoundFileImpactAllowsGeneratedDocChange(t *testing.T) {
	current := boundSurfaceImpactClaim()
	evidence, err := verifyClaimImpact("origin/main", current, current,
		map[string]bool{
			"internal/example.go": true,
			"docs/claim.md":       true,
			evidenceGraphManifest: true,
		})
	if err != nil {
		t.Fatalf("verify impact: %v", err)
	}
	if evidence.BoundSurfaceChangeCount != 1 ||
		len(evidence.BoundSurfaces[0].ChangedEvidence) != 1 ||
		len(evidence.BoundSurfaces[0].ChangedReasoningEvidence) != 1 {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
}

func TestBoundFileImpactRequiresReasoningEvidenceChange(t *testing.T) {
	current := boundSurfaceImpactClaim()
	if _, err := verifyClaimImpact("origin/main", current, current,
		map[string]bool{
			"internal/example.go": true,
			"docs/claim.md":       true,
		}); err == nil {
		t.Fatal("expected bound file change without reasoning evidence to fail")
	}
}

func boundSurfaceImpactClaim() []claimBinding {
	return []claimBinding{{
		ID: "claim", Statement: "kept", Loop: "loop",
		Files: []string{"internal/example.go"}, GeneratedDoc: []string{"docs/claim.md"},
	}}
}
