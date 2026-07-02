package main

import (
	"encoding/json"
	"os"
	"testing"
)

const notionBackfillFollowupReviewEvidence = "docs/30-architecture/evidence/notion-backfill-followup-review-2026-07-02.json"

func TestOperationalReadinessBindsNotionBackfillFollowupReview(t *testing.T) {
	check := readinessCheckByID(t, "staging_client_p0_visual_retest")
	if !hasMeasurement(check, "notion_backfill_followup_review_2026_07_02") {
		t.Fatal("missing Notion backfill follow-up review measurement")
	}
	if !hasEvidenceRef(check, notionBackfillFollowupReviewEvidence) {
		t.Fatal("missing Notion backfill follow-up evidence ref")
	}
}

func TestNotionBackfillFollowupSeparatesCWOpenWork(t *testing.T) {
	body, err := os.ReadFile("../../" + notionBackfillFollowupReviewEvidence)
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		Redacted        bool             `json:"redacted"`
		ReviewedChanges []reviewedSignal `json:"reviewed_changes"`
		CW              struct {
			Confirmed                  []string `json:"confirmed"`
			PolicyRequired             []string `json:"policy_required"`
			ClientDesignVisualRequired []string `json:"client_design_visual_required"`
			BackendBlocker             bool     `json:"backend_blocker"`
		} `json:"cw_structured_summary"`
	}
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	if !evidence.Redacted {
		t.Fatal("Notion follow-up evidence must stay redacted")
	}
	if len(evidence.CW.Confirmed) != 3 || len(evidence.CW.PolicyRequired) != 2 || len(evidence.CW.ClientDesignVisualRequired) != 6 {
		t.Fatalf("unexpected CW summary: %+v", evidence.CW)
	}
	if evidence.CW.BackendBlocker {
		t.Fatal("CW follow-up must not invent a backend blocker")
	}
	if !hasReviewedSignal(evidence.ReviewedChanges, "parent_body_still_contains_stale_dmg_workaround", "partial") {
		t.Fatal("missing stale parent body partial")
	}
	if !hasReviewedSignal(evidence.ReviewedChanges, "cw_open_items_are_policy_or_client_visual", "partial") {
		t.Fatal("missing CW policy/client visual partial")
	}
}
