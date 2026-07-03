package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestNotionQACommentReadbackLimitIsRecorded(t *testing.T) {
	evidence := loadNotionQAReadbackEvidence(t)
	if len(evidence.ReadbackObservations) == 0 {
		t.Fatal("missing Notion comment readback observations")
	}
	observation := evidence.ReadbackObservations[len(evidence.ReadbackObservations)-1]
	if observation.ObservedAt != "2026-07-03T08:13:54Z" {
		t.Fatalf("latest readback observation = %q", observation.ObservedAt)
	}
	if observation.VisiblePageLevelCommentCount != 31 {
		t.Fatalf("visible page-level count = %d", observation.VisiblePageLevelCommentCount)
	}
	if observation.LatestVisibleCommentID == "" {
		t.Fatal("latest visible comment id must be recorded")
	}
	for _, id := range []string{
		"39120241-cf7f-8172-b716-001dd440d1e1",
		"39120241-cf7f-818e-87dd-001d714d66ed",
		"39120241-cf7f-81c8-bd89-001d768e3b97",
		"39220241-cf7f-81ad-84b5-001df57331e5",
		"39220241-cf7f-8191-a0b4-001d15db7b78",
		"39220241-cf7f-8134-bc4e-001d98e7d805",
		"39220241-cf7f-8129-b12d-001d5b689249",
		"39220241-cf7f-81df-a487-001d7b84713a",
	} {
		if !hasReadbackMissingID(observation, id) {
			t.Fatalf("missing write-confirmed comment readback gap id=%s", id)
		}
	}
	if observation.Decision == "" {
		t.Fatal("readback gap decision must be explicit")
	}
}

func loadNotionQAReadbackEvidence(t *testing.T) notionQAReadbackEvidence {
	t.Helper()
	body, err := os.ReadFile("../../" + notionQACommentEventDiffEvidence)
	if err != nil {
		t.Fatal(err)
	}
	var evidence notionQAReadbackEvidence
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	return evidence
}

func hasReadbackMissingID(observation notionReadbackObservation, id string) bool {
	for _, got := range observation.WriteConfirmedNotVisible {
		if got == id {
			return true
		}
	}
	return false
}

type notionQAReadbackEvidence struct {
	ReadbackObservations []notionReadbackObservation `json:"readback_observations"`
}

type notionReadbackObservation struct {
	ObservedAt                   string   `json:"observed_at"`
	VisiblePageLevelCommentCount int      `json:"visible_page_level_comment_count"`
	LatestVisibleCommentID       string   `json:"latest_visible_comment_id"`
	WriteConfirmedNotVisible     []string `json:"write_confirmed_comment_ids_not_visible"`
	Decision                     string   `json:"decision"`
}
