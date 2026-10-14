package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestNotionQACommentEventsPreserveConcreteSignals(t *testing.T) {
	evidence := loadNotionQACommentEventDiff(t)
	if !evidence.Redacted {
		t.Fatal("Notion QA comment event diff evidence must stay redacted")
	}
	if len(evidence.CommentEvents) < 5 {
		t.Fatalf("comment event diff should preserve concrete comment events, got %d", len(evidence.CommentEvents))
	}
	wantEvents := map[string]string{
		"queued_by_busy_agent":               "server_fix_evidence",
		"waiting_for_user_intent":            "server_fix_evidence",
		"author_fallback":                    "contract_evidence",
		"codex_unknown_variant":              "daemon_release_evidence",
		"completion_progress_after_terminal": "evidence_request",
	}
	for discussion, eventType := range wantEvents {
		if !hasCommentEvent(evidence.CommentEvents, discussion, eventType) {
			t.Fatalf("missing comment event discussion=%s type=%s", discussion, eventType)
		}
	}
	if !containsString(evidence.DiffSummary.NewOrUnderBackfilled, "completion_progress_after_terminal") {
		t.Fatal("completion/progress terminal drift must be tracked as a separate gap")
	}
}

func loadNotionQACommentEventDiff(t *testing.T) notionQACommentEventDiff {
	t.Helper()
	body, err := os.ReadFile("../../" + notionQACommentEventDiffEvidence)
	if err != nil {
		t.Fatal(err)
	}
	var evidence notionQACommentEventDiff
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	return evidence
}

type notionQACommentEventDiff struct {
	Redacted      bool                 `json:"redacted"`
	CommentEvents []notionCommentEvent `json:"comment_events"`
	DiffSummary   struct {
		NewOrUnderBackfilled []string `json:"new_or_under_backfilled_items"`
	} `json:"diff_summary"`
}

type notionCommentEvent struct {
	ID         string `json:"id"`
	Discussion string `json:"discussion"`
	EventType  string `json:"event_type"`
	StatusTag  string `json:"status_tag"`
	Decision   string `json:"decision"`
}

func hasCommentEvent(events []notionCommentEvent, discussion, eventType string) bool {
	for _, event := range events {
		if event.Discussion != discussion || event.EventType != eventType {
			continue
		}
		return event.ID != "" && event.StatusTag != "" && event.Decision != ""
	}
	return false
}
