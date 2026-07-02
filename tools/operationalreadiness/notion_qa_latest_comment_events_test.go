package main

import "testing"

func TestNotionQACommentEventsIncludeLatestBackfillDiff(t *testing.T) {
	evidence := loadNotionQACommentEventDiff(t)
	want := map[string]string{
		"39120241-cf7f-81d5-9c0f-001d26c153a0": "queued_by_busy_agent",
		"39120241-cf7f-8119-b377-001d66282634": "completion_progress_after_terminal",
		"39120241-cf7f-816e-9523-001d6a184fdc": "page_level_backfill_diff",
	}
	for id, discussion := range want {
		if !hasCommentEventID(evidence.CommentEvents, id, discussion) {
			t.Fatalf("missing latest Notion comment event id=%s discussion=%s", id, discussion)
		}
	}
}

func hasCommentEventID(events []notionCommentEvent, id, discussion string) bool {
	for _, event := range events {
		if event.ID == id && event.Discussion == discussion {
			return event.EventType != "" && event.StatusTag != "" && event.Decision != ""
		}
	}
	return false
}
