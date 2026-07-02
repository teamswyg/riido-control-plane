package main

import (
	"strings"
	"testing"
)

const completionProgressCaptureToolEvidence = "docs/30-architecture/evidence/completion-progress-same-moment-capture-tool-2026-07-02.json"

func TestCompletionProgressCycleHasExecutableCaptureTool(t *testing.T) {
	m := loadManifestForTest(t)
	cycle := findNotionCycle(t, m, "notion_p0_completion_progress_after_terminal")
	if !notionCycleHasEvidenceRef(cycle, evidenceRef{Kind: "artifact", Path: completionProgressCaptureToolEvidence}) {
		t.Fatal("completion progress cycle must bind the same-moment capture tool evidence")
	}
	if !strings.Contains(cycle.RequiredNextCommand, "go run ./tools/aiagentthreadsnapshot") {
		t.Fatalf("completion progress next command is not executable capture tool: %s", cycle.RequiredNextCommand)
	}
	if strings.Contains(cycle.RequiredNextCommand, "<") || strings.Contains(cycle.RequiredNextCommand, "...") {
		t.Fatalf("completion progress next command must not contain placeholders: %s", cycle.RequiredNextCommand)
	}
	check := readinessCheckByID(t, "staging_client_p0_visual_retest")
	if !hasMeasurement(check, "completion_progress_same_moment_capture_tool_2026_07_02") {
		t.Fatal("visual retest check missing same-moment capture tool measurement")
	}
}
