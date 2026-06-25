package main

import "testing"

func TestClaimCoverageGapsReportMissingDimensions(t *testing.T) {
	deps := claimCoverageGapDeps()
	gaps := claimCoverageGaps(deps)
	if len(gaps) != 1 || gaps[0].ClaimID != "claim-a" {
		t.Fatalf("gaps = %+v", gaps)
	}
	if got := gaps[0].MissingDimensions; len(got) != 3 {
		t.Fatalf("missing dimensions = %+v", got)
	}
}

func TestLoopClosureAuditEvidenceExposesSyntheticClaimCoverageGaps(t *testing.T) {
	got := newEvidence(manifest{}, claimCoverageGapDeps())
	if got.ClaimCoverageGapCount != 1 || len(got.ClaimCoverageGaps) != 1 {
		t.Fatalf("expected synthetic claim coverage gap in evidence, got %+v", got)
	}
	first := got.ClaimCoverageGaps[0]
	if first.ClaimID == "" || first.Loop == "" || len(first.MissingDimensions) == 0 {
		t.Fatalf("incomplete claim coverage gap = %+v", first)
	}
}

func TestLoopClosureAuditManifestHasNoClaimCoverageGaps(t *testing.T) {
	m, deps := loadForTest(t)
	got := newEvidence(m, deps)
	if got.ClaimCoverageGapCount != 0 || len(got.ClaimCoverageGaps) != 0 {
		t.Fatalf("expected manifest claim coverage gaps to be closed, got %+v", got.ClaimCoverageGaps)
	}
}

func claimCoverageGapDeps() dependencies {
	return dependencies{registry: loopRegistry{
		Loops: []registryLoop{{
			ID:        "loop-a",
			Observes:  []string{"observation"},
			Verifies:  []string{"verification"},
			FailsWhen: []string{"failure"},
		}},
		Claims: []registryClaim{{ID: "claim-a", Loop: "loop-a"}},
	}}
}
