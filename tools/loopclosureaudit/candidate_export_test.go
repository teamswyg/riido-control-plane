package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestLoopClosureAuditExportsAuditGapCandidates(t *testing.T) {
	t.Setenv("RIIDO_LOOP_CLOSURE_AUDIT_NOW", "2026-06-24T12:00:00Z")
	m, deps := loadForTest(t)
	out := t.TempDir() + "/candidates.json"
	if err := run(options{Repo: "../..", Manifest: defaultManifest, CandidateOut: out}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var got candidateEvidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != candidateSchema || got.CandidateCount != len(got.Candidates) {
		t.Fatalf("candidate artifact shape = %+v", got)
	}
	coverageGaps := claimCoverageGaps(deps)
	want := len(m.ResidualGaps) + len(coverageGaps)
	if got.CandidateCount != want {
		t.Fatalf("candidate_count = %d, want %d", got.CandidateCount, want)
	}
	if got.CandidateCount > 0 && got.Candidates[0].PromotionTarget != "closed_loop_candidate" {
		t.Fatalf("candidate artifact content = %+v", got)
	}
	if len(coverageGaps) > 0 {
		if !hasCandidatePrefix(got.Candidates, "loop-closure-audit:claim_coverage:") {
			t.Fatalf("candidate artifact missing claim coverage candidates = %+v", got.Candidates)
		}
		if !hasClaimCoverageSubject(got.Candidates) {
			t.Fatalf("candidate artifact missing structured claim coverage subject = %+v", got.Candidates)
		}
	}
	if got.LiveStatus != candidateLiveStatus {
		t.Fatalf("live_status = %q", got.LiveStatus)
	}
}

func hasCandidatePrefix(candidates []closedLoopCandidate, prefix string) bool {
	for _, candidate := range candidates {
		if len(candidate.ID) >= len(prefix) && candidate.ID[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
