package main

import (
	"path/filepath"
	"testing"
)

func TestCandidateEvidenceFromIgnoredCommands(t *testing.T) {
	t.Setenv("GITHUB_RUN_ID", "run-ignored")
	plan := dispatchPlan{
		GeneratedAt: "2026-06-25T00:00:00Z",
		ExpiresAt:   "2026-06-26T00:00:00Z",
		IgnoredCommands: []selectedRefreshCommand{{
			LoopID:  "control_plane_pressure_candidate",
			Kind:    "target_verifier",
			Command: "go test ./tools/controlplaneperf",
		}},
	}
	got := candidateEvidenceFromPlan(plan)
	if got.CandidateCount != 1 || got.LiveStatus != "ignored_commands" {
		t.Fatalf("candidate evidence = %+v", got)
	}
	item := got.Candidates[0]
	if item.SourceRef.Run.ID != "run-ignored" || item.SourceRef.SummaryArtifact != dispatchSummaryArtifact {
		t.Fatalf("source_ref = %+v", item.SourceRef)
	}
	if item.AdoptionPlan[1].Command != "go test ./tools/controlplaneperf" {
		t.Fatalf("verifier command = %+v", item.AdoptionPlan)
	}
}

func TestLoopRefreshDispatchCLIWritesCandidates(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T00:00:00Z")
	dir := t.TempDir()
	in := filepath.Join(dir, "commands.json")
	dispatchOut := filepath.Join(dir, "dispatch.json")
	candidateOut := filepath.Join(dir, "candidates.json")
	writeIgnoredCommandInput(t, in)
	args := []string{
		"-repo", repoRootForTest(t),
		"-commands-in", in,
		"-dispatch-out", dispatchOut,
		"-candidate-out", candidateOut,
	}
	if err := mainRun(args); err != nil {
		t.Fatal(err)
	}
	if readCandidateEvidence(t, candidateOut).CandidateCount != 1 {
		t.Fatalf("candidate output missing ignored command")
	}
}

func writeIgnoredCommandInput(t *testing.T, path string) {
	t.Helper()
	err := writeJSON(path, refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema, Status: "refresh_required",
		GeneratedAt: "2026-06-24T00:00:00Z", ExpiresAt: "2026-06-26T00:00:00Z",
		CommandCount: 1,
		Commands: []selectedRefreshCommand{{
			LoopID: "closed_loop_candidate", Kind: "target_verifier",
			Command: "go test ./tools/looprefreshdispatch",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
}
