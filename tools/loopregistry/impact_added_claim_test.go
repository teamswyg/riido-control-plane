package main

import "testing"

func TestAddedClaimImpactRequiresBoundFileChange(t *testing.T) {
	current := []claimBinding{addedClaimImpactFixture()}
	if _, err := verifyClaimImpact("origin/main", nil, current,
		map[string]bool{"docs/claim.md": true, evidenceGraphManifest: true}); err == nil {
		t.Fatal("expected added claim without bound file change to fail")
	}
}

func TestAddedClaimImpactAllowsBoundFileAndReasoningChange(t *testing.T) {
	current := []claimBinding{addedClaimImpactFixture()}
	evidence, err := verifyClaimImpact("origin/main", nil, current,
		map[string]bool{"a.go": true, "docs/claim.md": true, evidenceGraphManifest: true})
	if err != nil {
		t.Fatalf("verify impact: %v", err)
	}
	if evidence.AddedClaimCount != 1 ||
		len(evidence.AddedClaims[0].ChangedEvidence) != 1 ||
		len(evidence.AddedClaims[0].ChangedReasoningEvidence) != 1 {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
}

func addedClaimImpactFixture() claimBinding {
	return claimBinding{
		ID:           "claim",
		Statement:    "new",
		Loop:         "loop",
		Files:        []string{"a.go"},
		GeneratedDoc: []string{"docs/claim.md"},
	}
}
