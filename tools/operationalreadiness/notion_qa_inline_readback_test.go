package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestNotionQAInlineDiscussionReadbackIsRecorded(t *testing.T) {
	body, err := os.ReadFile("../../" + notionQACommentEventDiffEvidence)
	if err != nil {
		t.Fatal(err)
	}
	var evidence notionQAInlineReadbackEvidence
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	if len(evidence.InlineReadbacks) == 0 {
		t.Fatal("missing targeted inline discussion readback evidence")
	}
	latest := evidence.InlineReadbacks[len(evidence.InlineReadbacks)-1]
	if latest.ObservedAt != "2026-07-02T22:21:07Z" {
		t.Fatalf("latest inline readback observed_at = %q", latest.ObservedAt)
	}
	want := map[string]string{
		"queued_by_busy_agent":               "39120241-cf7f-81d5-9c0f-001d26c153a0",
		"waiting_for_user_intent":            "39120241-cf7f-816a-8c1f-001d780c7b57",
		"multi_assignment_projection":        "39120241-cf7f-81e4-9f6e-001d6a501b66",
		"author_fallback":                    "39120241-cf7f-81e9-8e8b-001dfae5d349",
		"codex_unknown_variant":              "39120241-cf7f-8178-93e6-001d7d46aec8",
		"completion_progress_after_terminal": "39120241-cf7f-8130-b9a6-001d9e29226a",
	}
	for discussion, latestID := range want {
		if !hasInlineReadback(latest.Discussions, discussion, latestID) {
			t.Fatalf("missing inline discussion readback %s latest=%s", discussion, latestID)
		}
	}
	if latest.Decision == "" {
		t.Fatal("inline discussion readback must record a decision")
	}
}

func hasInlineReadback(readbacks []notionQAInlineDiscussionReadback, discussion, latestID string) bool {
	for _, readback := range readbacks {
		if readback.Discussion == discussion && readback.LatestVisibleCommentID == latestID {
			return readback.CommentCount > 0 && readback.LatestVisibleObservedAt != ""
		}
	}
	return false
}

type notionQAInlineReadbackEvidence struct {
	InlineReadbacks []struct {
		ObservedAt  string                             `json:"observed_at"`
		Decision    string                             `json:"decision"`
		Discussions []notionQAInlineDiscussionReadback `json:"discussions"`
	} `json:"inline_discussion_readbacks"`
}

type notionQAInlineDiscussionReadback struct {
	Discussion              string `json:"discussion"`
	CommentCount            int    `json:"comment_count"`
	LatestVisibleCommentID  string `json:"latest_visible_comment_id"`
	LatestVisibleObservedAt string `json:"latest_visible_comment_observed_at"`
}
