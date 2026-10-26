package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

type meetingReadCase struct {
	name, status                                            string
	deleted, denied, providerDenied, empty, providerMissing bool
}

func meetingReadUpstream(t *testing.T, tc meetingReadCase, calls *atomic.Int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer synthetic-requester" {
			t.Error("requester authority changed")
		}
		var req struct {
			Query     string
			Variables map[string]string
		}
		if json.NewDecoder(r.Body).Decode(&req) != nil || req.Variables["team"] != "team" || req.Variables["id"] != "meeting-1" {
			t.Error("incorrect meeting/team identity")
		}
		calls.Add(1)
		if strings.Contains(req.Query, "getMeeting(") {
			if tc.denied {
				writeJSON(w, 200, map[string]any{"errors": []any{map[string]any{"extensions": map[string]string{"code": "FORBIDDEN"}}}})
				return
			}
			var deleted any
			if tc.deleted {
				deleted = "2026-10-27T00:00:00Z"
			}
			writeJSON(w, 200, map[string]any{"data": map[string]any{"meeting": map[string]any{"id": "meeting-1", "deletedAt": deleted, "updatedAt": "2026-10-27T00:00:00Z", "document": map[string]any{"id": "doc-1", "content": meetingStoredFixtureContent(tc)}}}})
			return
		}
		if tc.providerDenied {
			writeJSON(w, 403, map[string]string{"message": "forbidden"})
			return
		}
		if tc.providerMissing {
			writeJSON(w, 200, map[string]any{"data": map[string]any{"structured": map[string]any{"tiptapJson": nil}, "rendered": map[string]any{"html": nil}}})
			return
		}
		doc := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"REQUESTER_MEETING_BODY"}]}]}`
		if tc.empty {
			doc = `{"type":"doc","content":[]}`
		}
		writeJSON(w, 200, map[string]any{"data": map[string]any{"structured": map[string]any{"tiptapJson": doc}, "rendered": map[string]any{"html": "<p>old fallback</p>"}}})
	}))
}

func meetingStoredFixtureContent(tc meetingReadCase) string {
	if tc.providerMissing {
		return ""
	}
	return "<p>stored older body</p>"
}
