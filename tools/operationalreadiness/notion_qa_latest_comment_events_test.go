package main

import "testing"

func TestNotionQACommentEventsIncludeLatestBackfillDiff(t *testing.T) {
	evidence := loadNotionQACommentEventDiff(t)
	want := map[string]string{
		"39120241-cf7f-81d5-9c0f-001d26c153a0": "queued_by_busy_agent",
		"39120241-cf7f-80fd-bc93-001d6a27203d": "queued_by_busy_agent",
		"39120241-cf7f-811c-acba-001d6b02996a": "queued_by_busy_agent",
		"39120241-cf7f-8149-ba94-001d3c2ecfa9": "queued_by_busy_agent",
		"39220241-cf7f-81a3-a228-001dcc9f4141": "queued_by_busy_agent",
		"39220241-cf7f-81e7-8cf7-001d57a1fc20": "queued_by_busy_agent",
		"39220241-cf7f-8114-8920-001d7b721670": "queued_by_busy_agent",
		"39120241-cf7f-81e4-9f6e-001d6a501b66": "multi_assignment_projection",
		"39120241-cf7f-816a-8c1f-001d780c7b57": "waiting_for_user_intent",
		"39120241-cf7f-81a6-b6b5-001d1b49303b": "codex_unknown_variant",
		"39120241-cf7f-8178-93e6-001d7d46aec8": "codex_unknown_variant",
		"39120241-cf7f-8119-b377-001d66282634": "completion_progress_after_terminal",
		"39120241-cf7f-81fe-8a5f-001d4070221f": "completion_progress_after_terminal",
		"39120241-cf7f-813f-aa9b-001dbe8f3a47": "completion_progress_after_terminal",
		"39120241-cf7f-81f1-bcdd-001da708227a": "completion_progress_after_terminal",
		"39120241-cf7f-8112-a42f-001d2d6ed4ec": "completion_progress_after_terminal",
		"39120241-cf7f-8137-9a8b-001d7a5d1615": "completion_progress_after_terminal",
		"39120241-cf7f-8130-b9a6-001d9e29226a": "completion_progress_after_terminal",
		"39120241-cf7f-816e-9523-001d6a184fdc": "page_level_backfill_diff",
		"39120241-cf7f-8148-bd04-001d6cbae868": "page_level_backfill_diff",
		"39120241-cf7f-8172-b716-001dd440d1e1": "page_level_backfill_diff",
		"39120241-cf7f-81c8-bd89-001d768e3b97": "page_level_backfill_diff",
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
