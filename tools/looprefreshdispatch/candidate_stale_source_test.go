package main

import (
	"path/filepath"
	"testing"
)

func TestLoopRefreshDispatchCLIWritesStaleSourceCandidate(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-29T00:00:00Z")
	dir := t.TempDir()
	in := filepath.Join(dir, "stale.json")
	dispatchOut := filepath.Join(dir, "dispatch.json")
	candidateOut := filepath.Join(dir, "candidates.json")
	if err := writeJSON(in, staleDecisionCommandSource()); err != nil {
		t.Fatal(err)
	}
	err := mainRun([]string{
		"-repo", repoRootForTest(t),
		"-commands-in", in,
		"-dispatch-out", dispatchOut,
		"-candidate-out", candidateOut,
	})
	if err != nil {
		t.Fatal(err)
	}
	plan := readDispatchPlan(t, dispatchOut)
	if plan.Status != "source_stale" || plan.SourceStaleCount != 1 {
		t.Fatalf("dispatch plan = %+v", plan)
	}
	candidates := readCandidateEvidence(t, candidateOut)
	if candidates.LiveStatus != "stale_sources" || candidates.CandidateCount != 1 {
		t.Fatalf("candidate evidence = %+v", candidates)
	}
}
