package main

import (
	"encoding/json"
	"os"
	"testing"
)

const notionQACommentEventDiffEvidence = "docs/30-architecture/evidence/notion-qa-comment-event-diff-2026-07-02.json"

func TestOperationalReadinessBindsNotionQACommentEventDiff(t *testing.T) {
	check := readinessCheckByID(t, "staging_client_p0_visual_retest")
	if !hasMeasurement(check, "notion_qa_comment_event_diff_2026_07_02") {
		t.Fatal("missing Notion QA comment event diff measurement")
	}
	if !hasEvidenceRef(check, notionQACommentEventDiffEvidence) {
		t.Fatal("missing Notion QA comment event diff evidence ref")
	}
	cycle := findNotionCycle(t, loadManifestForTest(t), "notion_p0_completion_progress_after_terminal")
	if cycle.CodexStatus != "[codex][진행중][부분증적]" {
		t.Fatalf("completion/progress codex_status = %q", cycle.CodexStatus)
	}
	if cycle.RequiredNextArtifact != "reproduced_completion_progress_same_conversation_network_snapshot" {
		t.Fatalf("completion/progress next artifact = %q", cycle.RequiredNextArtifact)
	}
	want := evidenceRef{Kind: "external", Path: "notion-comment:39120241-cf7f-81ec-99e9-001d90f2fa61"}
	if !notionCycleHasEvidenceRef(cycle, want) {
		t.Fatalf("completion/progress cycle missing Notion evidence ref %+v", want)
	}
	baselineComment := evidenceRef{Kind: "external", Path: "notion-comment:39120241-cf7f-8196-a65c-001d664f0e7a"}
	if !notionCycleHasEvidenceRef(cycle, baselineComment) {
		t.Fatalf("completion/progress cycle missing baseline comment ref %+v", baselineComment)
	}
	api := evidenceRef{Kind: "api_snapshot", Path: stagingCompletionProgressReadOnlyEvidence}
	if !notionCycleHasEvidenceRef(cycle, api) {
		t.Fatalf("completion/progress cycle missing read-only baseline %+v", api)
	}
}

func TestNotionQACommentEventDiffSeparatesNewCompletionProgressGap(t *testing.T) {
	body, err := os.ReadFile("../../" + notionQACommentEventDiffEvidence)
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		Redacted    bool `json:"redacted"`
		DiffSummary struct {
			NewOrUnderBackfilled []string `json:"new_or_under_backfilled_items"`
		} `json:"diff_summary"`
	}
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	if !evidence.Redacted {
		t.Fatal("Notion QA comment event diff evidence must stay redacted")
	}
	if !containsString(evidence.DiffSummary.NewOrUnderBackfilled, "completion_progress_after_terminal") {
		t.Fatal("completion/progress terminal drift must be tracked as a separate gap")
	}
}
