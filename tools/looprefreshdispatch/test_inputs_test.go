package main

import "testing"

func writeWorkflowCommandInput(t *testing.T, path string) {
	t.Helper()
	err := writeJSON(path, refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
		Status:        "refresh_required",
		GeneratedAt:   "2026-06-24T00:00:00Z",
		ExpiresAt:     "2026-06-26T00:00:00Z",
		CommandCount:  1,
		Commands: []selectedRefreshCommand{{
			LoopID:  "closed_loop_candidate",
			Kind:    "refresh_workflow",
			Command: "gh workflow run closed-loop-candidate-intake.yml --ref main",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
}
