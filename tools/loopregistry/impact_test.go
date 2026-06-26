package main

import "testing"

func TestClaimImpactRequiresBoundFileChange(t *testing.T) {
	base := []claimBinding{{
		ID:           "claim",
		Statement:    "old",
		Loop:         "loop",
		Files:        []string{"internal/example.go"},
		GeneratedDoc: []string{"docs/claim.md"},
	}}
	current := []claimBinding{{
		ID:           "claim",
		Statement:    "new",
		Loop:         "loop",
		Files:        []string{"internal/example.go"},
		GeneratedDoc: []string{"docs/claim.md"},
	}}
	if _, err := verifyClaimImpact("origin/main", base, current, map[string]bool{}); err == nil {
		t.Fatal("expected missing bound file change to fail")
	}
}

func TestClaimImpactAllowsBoundFileChange(t *testing.T) {
	base := []claimBinding{impactClaimFixture("old")}
	current := []claimBinding{impactClaimFixture("new")}
	evidence, err := verifyClaimImpact("origin/main", base, current, map[string]bool{
		"a.go": true, "docs/claim.md": true, evidenceGraphManifest: true,
	})
	if err != nil {
		t.Fatalf("verify impact: %v", err)
	}
	if evidence.ChangedClaimCount != 1 ||
		len(evidence.Claims[0].ChangedBoundFiles) != 1 ||
		len(evidence.Claims[0].ChangedEvidence) != 1 ||
		len(evidence.Claims[0].ChangedReasoningEvidence) != 1 {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
}

func TestClaimImpactRequiresGeneratedEvidenceForMeaningChange(t *testing.T) {
	base := []claimBinding{impactClaimFixture("old")}
	current := []claimBinding{impactClaimFixture("new")}
	if _, err := verifyClaimImpact("origin/main", base, current, map[string]bool{
		"a.go": true, evidenceGraphManifest: true,
	}); err == nil {
		t.Fatal("expected missing generated claim evidence change to fail")
	}
}

func TestClaimImpactRequiresEvidenceGraphReasoningForMeaningChange(t *testing.T) {
	base := []claimBinding{impactClaimFixture("old")}
	current := []claimBinding{impactClaimFixture("new")}
	if _, err := verifyClaimImpact("origin/main", base, current, map[string]bool{
		"a.go": true, "docs/claim.md": true,
	}); err == nil {
		t.Fatal("expected missing evidence graph reasoning change to fail")
	}
}

func impactClaimFixture(statement string) claimBinding {
	return claimBinding{
		ID:           "claim",
		Statement:    statement,
		Loop:         "loop",
		Files:        []string{"a.go"},
		GeneratedDoc: []string{"docs/claim.md"},
	}
}
