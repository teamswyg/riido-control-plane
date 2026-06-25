package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestLoopClosureAuditExportsResidualGapCandidates(t *testing.T) {
	t.Setenv("RIIDO_LOOP_CLOSURE_AUDIT_NOW", "2026-06-24T12:00:00Z")
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
	if got.CandidateCount < 2 || got.Candidates[0].PromotionTarget != "closed_loop_candidate" {
		t.Fatalf("candidate artifact content = %+v", got)
	}
}
