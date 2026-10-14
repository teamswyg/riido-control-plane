package main

import "testing"

func TestRemovedClaimImpactRequiresBoundFileChange(t *testing.T) {
	base := []claimBinding{removedClaimImpactFixture()}
	if _, err := verifyClaimImpact("origin/main", base, nil,
		map[string]bool{"docs/claim.md": true, evidenceGraphManifest: true}); err == nil {
		t.Fatal("expected removed claim without bound file change to fail")
	}
}

func TestRemovedClaimImpactAllowsBoundFileAndReasoningChange(t *testing.T) {
	base := []claimBinding{removedClaimImpactFixture()}
	evidence, err := verifyClaimImpact("origin/main", base, nil,
		map[string]bool{"a.go": true, "docs/claim.md": true, evidenceGraphManifest: true})
	if err != nil {
		t.Fatalf("verify impact: %v", err)
	}
	if evidence.RemovedClaimCount != 1 ||
		len(evidence.RemovedClaims[0].ChangedEvidence) != 1 ||
		len(evidence.RemovedClaims[0].ChangedReasoningEvidence) != 1 {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
}

func removedClaimImpactFixture() claimBinding {
	return claimBinding{
		ID:           "claim",
		Statement:    "old",
		Loop:         "loop",
		Files:        []string{"a.go"},
		GeneratedDoc: []string{"docs/claim.md"},
	}
}
