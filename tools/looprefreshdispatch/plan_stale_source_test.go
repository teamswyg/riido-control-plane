package main

import "testing"

func TestBuildDispatchPlanReportsSkippedStaleSources(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-29T00:00:00Z")
	source, err := mergeRefreshCommandSources([]refreshCommandEvidence{
		freshWorkflowCommandSource(),
		staleDecisionCommandSource(),
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := buildDispatchPlan(repoRootForTest(t), source)
	if err != nil {
		t.Fatal(err)
	}
	if got.SourceStaleCount != 1 ||
		got.SourceStaleSources[0].SourcePath != "stale-decision.json" {
		t.Fatalf("dispatch stale source evidence = %+v", got)
	}
}
