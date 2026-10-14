package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestNotionQAReadbackGapSourcesAreExplained(t *testing.T) {
	evidence := loadNotionQAReadbackSourceEvidence(t)
	if len(evidence.ReadbackObservations) == 0 {
		t.Fatal("missing readback observations")
	}
	observation := evidence.ReadbackObservations[len(evidence.ReadbackObservations)-1]
	for _, id := range observation.WriteConfirmedNotVisible {
		source, ok := findReadbackSource(observation.Sources, id)
		if !ok {
			t.Fatalf("missing readback source for write-confirmed id %s", id)
		}
		if source.CommentEventRequired && !hasCommentEventAny(evidence.CommentEvents, id) {
			t.Fatalf("missing required comment_event for write-confirmed id %s", id)
		}
		if !source.CommentEventRequired && source.Reason == "" {
			t.Fatalf("gap-only readback id %s must explain why no comment_event is required", id)
		}
	}
}

func loadNotionQAReadbackSourceEvidence(t *testing.T) notionQAReadbackSourceEvidence {
	t.Helper()
	var sourceEvidence notionQAReadbackSourceEvidence
	body, err := os.ReadFile("../../" + notionQACommentEventDiffEvidence)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &sourceEvidence); err != nil {
		t.Fatal(err)
	}
	return sourceEvidence
}

func findReadbackSource(sources []notionReadbackSource, id string) (notionReadbackSource, bool) {
	for _, source := range sources {
		if source.ID == id {
			return source, true
		}
	}
	return notionReadbackSource{}, false
}

func hasCommentEventAny(events []notionCommentEvent, id string) bool {
	for _, event := range events {
		if event.ID == id {
			return event.EventType != "" && event.StatusTag != "" && event.Decision != ""
		}
	}
	return false
}

type notionQAReadbackSourceEvidence struct {
	CommentEvents        []notionCommentEvent      `json:"comment_events"`
	ReadbackObservations []notionSourceObservation `json:"readback_observations"`
}

type notionSourceObservation struct {
	WriteConfirmedNotVisible []string               `json:"write_confirmed_comment_ids_not_visible"`
	Sources                  []notionReadbackSource `json:"write_confirmed_comment_sources"`
}

type notionReadbackSource struct {
	ID                   string `json:"id"`
	CommentEventRequired bool   `json:"comment_event_required"`
	Reason               string `json:"reason"`
}
