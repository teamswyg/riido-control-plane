package main

import "testing"

func TestLoopRefreshDispatchCLIAcceptsRepeatedCommandsIn(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T00:00:00Z")
	dir := t.TempDir()
	first := dir + "/first.json"
	second := dir + "/second.json"
	out := dir + "/dispatch.json"
	writeWorkflowCommandInput(t, first)
	writeIgnoredCommandInput(t, second)
	err := mainRun([]string{
		"-repo", repoRootForTest(t),
		"-commands-in", first,
		"-commands-in", second,
		"-dispatch-out", out,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := readDispatchPlan(t, out)
	if got.SourceCommandCount != 2 || got.DispatchCount != 1 || got.IgnoredCommandCount != 1 {
		t.Fatalf("merged dispatch plan = %+v", got)
	}
}

func TestMergeRefreshCommandsUsesIntersectionWindow(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T12:00:00Z")
	got, err := mergeRefreshCommandSources([]refreshCommandEvidence{{
		SchemaVersion: refreshCommandsSchema,
		Status:        "refresh_required",
		GeneratedAt:   "2026-06-24T00:00:00Z",
		ExpiresAt:     "2026-06-27T00:00:00Z",
		CommandCount:  1,
		Commands: []selectedRefreshCommand{{
			LoopID: "closed_loop_candidate", Kind: "target_verifier",
			Command: "go test ./tools/looprefreshdispatch",
		}},
	}, {
		SchemaVersion: refreshCommandsSchema,
		Status:        "fresh",
		GeneratedAt:   "2026-06-25T00:00:00Z",
		ExpiresAt:     "2026-06-26T00:00:00Z",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if got.GeneratedAt != "2026-06-25T00:00:00Z" || got.ExpiresAt != "2026-06-26T00:00:00Z" {
		t.Fatalf("merged freshness window = %+v", got)
	}
}
