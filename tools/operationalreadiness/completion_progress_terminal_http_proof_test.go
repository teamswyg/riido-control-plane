package main

import "testing"

func TestOperationalReadinessBindsCompletionProgressCompletedTerminalProof(t *testing.T) {
	m := loadManifestForTest(t)
	cycle := findNotionCycle(t, m, "notion_p0_completion_progress_after_terminal")
	want := []evidenceRef{
		{
			Kind: "test",
			Path: "internal/riidoaiserver/ai_agent_client_http_v3_completed_refresh_test.go",
		},
		{
			Kind: "artifact",
			Path: "docs/30-architecture/evidence/completion-progress-completed-terminal-http-proof-2026-07-03.json",
		},
	}
	for _, ref := range want {
		if !notionCycleHasEvidenceRef(cycle, ref) {
			t.Fatalf("completion progress P0 cycle missing evidence ref %+v", ref)
		}
	}
	check := readinessCheckByID(t, "staging_client_p0_visual_retest")
	if !hasMeasurement(check, "completion_progress_completed_terminal_http_proof_2026_07_03") {
		t.Fatal("visual retest check missing completed-terminal HTTP proof measurement")
	}
}
