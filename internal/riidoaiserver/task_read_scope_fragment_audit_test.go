package riidoaiserver

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTaskReadScopeFragmentMetadataAuditsDeliveredLeavesOnly(t *testing.T) {
	raw, _ := json.Marshal(map[string]any{"id": "work", "kind": "component", "teamId": "team", "details": map[string]any{"description": strings.Repeat("x", 40<<10)}, "links": []any{map[string]string{"id": "later-link", "url": "https://example.test"}}, "github": []any{}, "attachments": []any{}})
	var work taskReadScopeWork
	if err := json.Unmarshal(raw, &work); err != nil {
		t.Fatal(err)
	}
	b, _ := newTaskReadScopeBuilder()
	part := b.captureWorkMetadata(AssignmentScopedResource{Kind: AssignmentContentComponent, ID: "work"}, work)
	if !part.Complete || len(b.capture.Pages) < 2 {
		t.Fatal("fragment capture incomplete")
	}
	for _, source := range b.capture.Pages[0].Sources {
		if source.Kind == "link" {
			t.Fatal("undelivered later link audited")
		}
	}
	found := false
	for _, page := range b.capture.Pages[1:] {
		for _, source := range page.Sources {
			found = found || source.ID == "later-link"
		}
	}
	if !found {
		t.Fatal("delivered link audit missing")
	}
}

func TestTaskReadScopeFragmentCommentsAuditDeliveredLeavesOnly(t *testing.T) {
	item := func(id, body string) json.RawMessage {
		raw, _ := json.Marshal(map[string]any{"id": id, "body": body, "replyCount": 0})
		return raw
	}
	complete := true
	first := &taskReadCommentPage{ResourceID: "work", ResourceKind: AssignmentContentComponent, Version: "comment-version", Status: "available", Page: 1, Limit: 20, Complete: &complete, Items: []json.RawMessage{item("first-comment", strings.Repeat("x", 40<<10)), item("later-comment", "later body")}}
	b, _ := newTaskReadScopeBuilder()
	c := &AIAgentPrivateTaskContextClient{}
	part, err := c.captureReadScopeComments(t.Context(), "synthetic", b, AssignmentScopedResource{Kind: AssignmentContentComponent, ID: "work", TeamID: "team"}, first)
	if err != nil || !part.Complete || len(b.capture.Pages) < 2 {
		t.Fatalf("comment fragments: %v", err)
	}
	for _, source := range b.capture.Pages[0].Sources {
		if source.ID == "later-comment" {
			t.Fatal("undelivered later comment audited")
		}
	}
	found := false
	for _, page := range b.capture.Pages[1:] {
		for _, source := range page.Sources {
			found = found || source.ID == "later-comment"
		}
	}
	if !found {
		t.Fatal("delivered comment audit missing")
	}
}
