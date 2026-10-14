package main

import (
	"path/filepath"
	"testing"
)

func TestLoopRefreshCommandFixtureDrivesDispatchAndCandidates(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T00:00:00Z")
	root := repoRootForTest(t)
	source, err := loadRefreshCommands(loopRefreshFixturePath(root))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := buildDispatchPlan(root, source)
	if err != nil {
		t.Fatal(err)
	}
	if plan.DispatchCount != 2 || plan.IgnoredCommandCount != 1 {
		t.Fatalf("fixture plan = %+v", plan)
	}
	candidates := candidateEvidenceFromPlan(plan)
	if candidates.CandidateCount != 1 {
		t.Fatalf("fixture candidates = %+v", candidates)
	}
}

func loopRefreshFixturePath(root string) string {
	return filepath.Join(
		root,
		"docs",
		"30-architecture",
		"fixtures",
		"loop-refresh-commands.fixture.json",
	)
}
