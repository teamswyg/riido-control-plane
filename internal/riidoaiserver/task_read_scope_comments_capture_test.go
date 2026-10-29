package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestTaskReadScopeCapturesCommentPagesAndRepliesWithoutLiveFallback(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var req struct{ Variables map[string]any }
		if json.NewDecoder(r.Body).Decode(&req) != nil {
			t.Error("invalid request")
		}
		page := readScopeCommentFixture("parent", 1, []any{map[string]any{"id": "reply", "body": "reply body", "replyCount": 0}}, nil, true)
		if req.Variables["parentCommentId"] == nil {
			if req.Variables["version"] != "root-comments" || req.Variables["page"] != float64(2) {
				t.Errorf("continuation lost version: %v", req.Variables)
			}
			page = readScopeCommentFixture("", 2, []any{map[string]any{"id": "second", "body": "second body", "replyCount": 0}}, nil, true)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getReadScopeComments": page}})
	}))
	t.Cleanup(server.Close)
	c, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL})
	firstJSON, _ := json.Marshal(readScopeCommentFixture("", 1, []any{map[string]any{"id": "parent", "body": "first body", "replyCount": 1}}, 2, false))
	var first taskReadCommentPage
	_ = json.Unmarshal(firstJSON, &first)
	b, _ := newTaskReadScopeBuilder()
	resource := AssignmentScopedResource{Kind: AssignmentContentComponent, ID: "work", TeamID: "team", Reasons: []taskReadScopeReason{{Kind: "root"}}}
	part, err := c.captureReadScopeComments(t.Context(), "requester", b, resource, &first)
	if err != nil || !part.Complete || len(part.Cursors) != 3 || calls.Load() != 2 {
		t.Fatalf("comments incomplete: %+v calls=%d err=%v", part, calls.Load(), err)
	}
	ids := map[string]bool{}
	for _, page := range b.capture.Pages {
		for _, source := range page.Sources {
			if source.Kind == "comment" {
				ids[source.ID] = true
			}
		}
	}
	if !ids["parent"] || !ids["second"] || !ids["reply"] {
		t.Fatalf("leaf audit missing: %v", ids)
	}
}

func readScopeCommentFixture(parent string, page int, items []any, next any, complete bool) map[string]any {
	var parentID any
	version := "root-comments"
	if parent != "" {
		parentID = parent
		version = "reply-comments"
	}
	status := "available"
	if !complete {
		status = "partial"
	}
	return map[string]any{"resourceId": "work", "resourceKind": "component", "parentCommentId": parentID, "version": version, "status": status, "page": page, "limit": 1, "nextPage": next, "complete": complete, "items": items}
}
