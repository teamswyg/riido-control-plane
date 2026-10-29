package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestTaskReadScopeMeetingCapturesAllBoardCardsAndEmptyGroups(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var req struct{ Variables map[string]any }
		if json.NewDecoder(r.Body).Decode(&req) != nil {
			t.Error("invalid request")
		}
		offset := 0
		var next any = "next-board"
		complete := false
		if req.Variables["boardCursor"] == "next-board" {
			offset = 1
			next = nil
			complete = true
		}
		body := map[string]any{"resourceId": "meeting", "part": "document", "status": "empty", "source": "snapshot_html", "format": "html", "version": readScopeBodyVersion(AssignmentContentMeeting, "meeting", "document", "snapshot_html", "html", ""), "offset": 0, "content": "", "nextOffset": nil, "complete": true}
		description := map[string]any{"resourceId": "meeting", "part": "description", "status": "empty", "source": "snapshot_description", "format": "html", "version": readScopeBodyVersion(AssignmentContentMeeting, "meeting", "description", "snapshot_description", "html", ""), "offset": 0, "content": "", "nextOffset": nil, "complete": true}
		meeting := map[string]any{"id": "meeting", "teamId": "team", "status": "available", "details": map[string]any{"id": "meeting", "title": "Planning", "meetingParticipants": []any{map[string]string{"id": "empty-user"}}}, "labels": []any{}, "board": map[string]any{"version": "board-version", "offset": offset, "nextCursor": next, "complete": complete, "groups": []any{map[string]string{"id": "empty-group"}}, "cards": []any{map[string]any{"id": []string{"card-a", "card-b"}[offset], "body": "<p>original</p>"}}}, "body": body, "descriptionBody": description, "partsStatus": []any{}}
		if !complete {
			meeting["status"] = "partial" // Actual producer aggregates incomplete board pagination.
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getMeetingReadScope": meeting}})
	}))
	t.Cleanup(server.Close)
	c, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL})
	b, _ := newTaskReadScopeBuilder()
	resource, err := c.captureReadScopeMeeting(t.Context(), "requester", b, taskReadScopeMeetingRef{ID: "meeting", TeamID: "team", Reasons: []taskReadScopeReason{{Kind: "meeting"}}})
	if err != nil || calls.Load() != 2 || len(resource.Parts) != 4 {
		t.Fatalf("meeting capture: %+v calls=%d err=%v", resource, calls.Load(), err)
	}
	for _, part := range resource.Parts {
		if !part.Complete {
			t.Fatalf("meeting part incomplete: %+v", part)
		}
	}
	encoded, _ := json.Marshal(b.capture.Pages)
	for _, want := range []string{"card-a", "card-b", "empty-group", "empty-user"} {
		if !strings.Contains(string(encoded), want) {
			t.Errorf("missing %s", want)
		}
	}
}
