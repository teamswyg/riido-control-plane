package main

import "testing"

func TestNotionQACommentEventsIncludeLatestBackfillDiff(t *testing.T) {
	evidence := loadNotionQACommentEventDiff(t)
	want := map[string]string{
		"39120241-cf7f-81d5-9c0f-001d26c153a0": "queued_by_busy_agent",
		"39120241-cf7f-8119-b377-001d66282634": "completion_progress_after_terminal",
		"39120241-cf7f-81fe-8a5f-001d4070221f": "completion_progress_after_terminal",
		"39120241-cf7f-813f-aa9b-001dbe8f3a47": "completion_progress_after_terminal",
		"39120241-cf7f-81f1-bcdd-001da708227a": "completion_progress_after_terminal",
		"39120241-cf7f-8112-a42f-001d2d6ed4ec": "completion_progress_after_terminal",
		"39120241-cf7f-8137-9a8b-001d7a5d1615": "completion_progress_after_terminal",
		"39120241-cf7f-8130-b9a6-001d9e29226a": "completion_progress_after_terminal",
		"39120241-cf7f-816e-9523-001d6a184fdc": "page_level_backfill_diff",
		"39120241-cf7f-8148-bd04-001d6cbae868": "page_level_backfill_diff",
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
