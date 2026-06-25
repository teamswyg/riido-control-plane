package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoopRefreshDispatchCLIWritesPlan(t *testing.T) {
	root := repoRootForTest(t)
	dir := t.TempDir()
	in := filepath.Join(dir, "commands.json")
	out := filepath.Join(dir, "dispatch.json")
	err := writeJSON(in, refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
		Status:        "refresh_required",
		Commands: []selectedRefreshCommand{{
			LoopID:  "closed_loop_candidate",
			Kind:    "refresh_workflow",
			Command: "gh workflow run closed-loop-candidate-intake.yml --ref main",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := mainRun([]string{"-repo", root, "-commands-in", in, "-dispatch-out", out}); err != nil {
		t.Fatal(err)
	}
	got := readDispatchPlan(t, out)
	if got.SchemaVersion != dispatchPlanSchema ||
		got.Dispatches[0].WorkflowFile != "closed-loop-candidate-intake.yml" {
		t.Fatalf("dispatch plan = %+v", got)
	}
}

func readDispatchPlan(t *testing.T, path string) dispatchPlan {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got dispatchPlan
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	return got
}

func repoRootForTest(t *testing.T) string {
	t.Helper()
	root, err := findRepoRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	return root
}
