package main

import "testing"

func TestHarnessPromotionManifestVerifies(t *testing.T) {
	if err := run(options{
		Repo:        "../..",
		Manifest:    defaultManifest,
		CheckDoc:    true,
		EvidenceOut: t.TempDir() + "/evidence.json",
	}); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestHarnessFailurePromotesUnverifiedClaims(t *testing.T) {
	source := promotionSource{
		ID: "smoke", HarnessLoop: "provider_acceptance_harness",
		PromotionTarget: "closed_loop_candidate", FailureStatuses: []string{"failure"},
		RequiredNextArtifacts: []string{"claim_binding", "verifier", "ci_gate"},
	}
	summary := liveSummary{ID: "smoke", LiveStatus: "failure", EvidenceClaims: []liveClaim{
		{ID: "ok", Status: "verified"}, {ID: "broken", Summary: "broken claim", Status: "not_verified"},
	}}
	got := buildCandidateEvidence(source, summary)
	if got.CandidateCount != 1 || got.Candidates[0].ID != "smoke:broken" {
		t.Fatalf("candidates = %+v", got.Candidates)
	}
}

func TestPromotionSourceMustBindLoopRegistryTarget(t *testing.T) {
	m := loadPromotionManifestForTest(t)
	m.Sources[0].PromotionTarget = "missing_loop"
	if _, err := verifyAll("../..", m); err == nil {
		t.Fatal("expected missing loop registry target to fail")
	}
}

func TestPromotionSourceMustRequireDecisionAndGraph(t *testing.T) {
	m := loadPromotionManifestForTest(t)
	m.Sources[0].RequiredNextArtifacts = []string{"claim_binding", "verifier", "ci_gate", "redacted_evidence"}
	if _, err := verifyAll("../..", m); err == nil {
		t.Fatal("expected missing adoption artifacts to fail")
	}
}

func loadPromotionManifestForTest(t *testing.T) manifest {
	t.Helper()
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	m, err := loadManifest(repoPath(root, defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	return m
}
