package main

import (
	"encoding/json"
	"os"
	"testing"
)

const notionBackfillReviewEvidence = "docs/30-architecture/evidence/notion-backfill-review-2026-07-02.json"

type reviewedSignal struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
}

func TestOperationalReadinessBindsNotionBackfillReview(t *testing.T) {
	check := readinessCheckByID(t, "staging_client_p0_visual_retest")
	if !hasMeasurement(check, "notion_backfill_review_2026_07_02") {
		t.Fatal("missing Notion backfill review measurement")
	}
	if !hasEvidenceRef(check, notionBackfillReviewEvidence) {
		t.Fatal("missing Notion backfill review evidence ref")
	}
}

func TestNotionBackfillReviewKeepsStaleBodyPartial(t *testing.T) {
	body, err := os.ReadFile("../../" + notionBackfillReviewEvidence)
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		Redacted        bool             `json:"redacted"`
		ReviewedSignals []reviewedSignal `json:"reviewed_signals"`
		RepoState       struct {
			NotionOpenLoopCycleCount int `json:"notion_open_loop_cycle_count"`
			NotionOpenLoopP0Count    int `json:"notion_open_loop_p0_count"`
		} `json:"repo_state"`
	}
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	if !evidence.Redacted {
		t.Fatal("Notion backfill evidence must stay redacted")
	}
	if evidence.RepoState.NotionOpenLoopCycleCount != 7 || evidence.RepoState.NotionOpenLoopP0Count != 7 {
		t.Fatalf("unexpected Notion open-loop counts: %+v", evidence.RepoState)
	}
	if !hasReviewedSignal(evidence.ReviewedSignals, "notion_body_gatekeeper_text_is_stale", "partial") {
		t.Fatal("missing stale Notion body partial signal")
	}
	if !hasReviewedSignal(evidence.ReviewedSignals, "p0_backend_fix_status_remains_visual_partial", "partial") {
		t.Fatal("missing visual-product partial signal")
	}
}

func hasReviewedSignal(signals []reviewedSignal, id, severity string) bool {
	for _, signal := range signals {
		if signal.ID == id && signal.Severity == severity {
			return true
		}
	}
	return false
}
