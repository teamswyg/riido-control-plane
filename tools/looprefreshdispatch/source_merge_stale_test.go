package main

import "testing"

func TestMergeRefreshCommandsSkipsStaleSourceWhenFreshSourceExists(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-29T00:00:00Z")
	got, err := mergeRefreshCommandSources([]refreshCommandEvidence{
		freshWorkflowCommandSource(),
		staleDecisionCommandSource(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.CommandCount != 1 || len(got.StaleSources) != 1 {
		t.Fatalf("merged source = %+v", got)
	}
	if got.StaleSources[0].SourcePath != "stale-decision.json" {
		t.Fatalf("stale source = %+v", got.StaleSources[0])
	}
}

func TestMergeRefreshCommandsReportsWhenAllSourcesAreStale(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-29T00:00:00Z")
	got, err := mergeRefreshCommandSources([]refreshCommandEvidence{
		staleDecisionCommandSource(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "source_stale" || len(got.StaleSources) != 1 {
		t.Fatalf("all-stale source evidence = %+v", got)
	}
}

func freshWorkflowCommandSource() refreshCommandEvidence {
	return refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
		Status:        "refresh_required",
		GeneratedAt:   "2026-06-28T00:00:00Z",
		ExpiresAt:     "2026-06-30T00:00:00Z",
		CommandCount:  1,
		Commands: []selectedRefreshCommand{{
			LoopID:  "loop_registry",
			Kind:    "refresh_workflow",
			Command: "gh workflow run loop-registry.yml --ref main",
		}},
	}
}

func staleDecisionCommandSource() refreshCommandEvidence {
	return refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
		Status:        "refresh_required",
		GeneratedAt:   "2026-06-24T00:00:00Z",
		ExpiresAt:     "2026-06-25T00:00:00Z",
		SourcePath:    "stale-decision.json",
		CommandCount:  1,
		Commands: []selectedRefreshCommand{{
			LoopID: "closed_loop_candidate", Kind: "target_verifier",
			Command: "go test ./tools/closedloopcandidatedecision",
		}},
	}
}
