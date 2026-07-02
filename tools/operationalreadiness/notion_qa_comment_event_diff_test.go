package main

import "testing"

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
	followupComment := evidenceRef{Kind: "external", Path: "notion-comment:39120241-cf7f-8125-bd7a-001d0eff72b8"}
	if !notionCycleHasEvidenceRef(cycle, followupComment) {
		t.Fatalf("completion/progress cycle missing follow-up comment ref %+v", followupComment)
	}
	api := evidenceRef{Kind: "api_snapshot", Path: stagingCompletionProgressReadOnlyEvidence}
	if !notionCycleHasEvidenceRef(cycle, api) {
		t.Fatalf("completion/progress cycle missing read-only baseline %+v", api)
	}
	followup := evidenceRef{Kind: "api_snapshot", Path: stagingCompletionProgressFollowupEvidence}
	if !notionCycleHasEvidenceRef(cycle, followup) {
		t.Fatalf("completion/progress cycle missing follow-up baseline %+v", followup)
	}
}
