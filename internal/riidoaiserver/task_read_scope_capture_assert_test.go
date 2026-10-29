package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func readScopeAssertBodyHTTP(t *testing.T, f *assignmentContextFixture, snapshot AssignmentContextSnapshot) {
	t.Helper()
	var body string
	for _, part := range snapshot.Content {
		if part.Part != AssignmentContentEditor {
			continue
		}
		if !part.Complete || len(part.Cursors) != 2 {
			t.Fatalf("body incomplete: %+v", part)
		}
		for i, cursor := range part.Cursors {
			req, _ := json.Marshal(AgentAssignmentContextRequest{AssignmentID: f.leased.ID, LeaseToken: f.leased.LeaseToken, View: "content", Target: "task-private", Cursor: cursor})
			status, encoded := f.do("/v1/agents/agent-owned-codex/assignment-context", string(req), f.device)
			if status != http.StatusOK {
				t.Fatalf("content HTTP status=%d body=%s", status, encoded)
			}
			var out struct {
				Data struct {
					Data struct {
						Content string `json:"content"`
					} `json:"data"`
					NextCursor string `json:"next_cursor"`
				} `json:"data"`
			}
			if json.Unmarshal([]byte(encoded), &out) != nil {
				t.Fatal("invalid content response")
			}
			body += out.Data.Data.Content
			if (i == 0) != (out.Data.NextCursor != "") {
				t.Fatal("capture next cursor mismatch")
			}
		}
	}
	if body != "helloworld" {
		t.Fatalf("full captured body=%q", body)
	}
}
