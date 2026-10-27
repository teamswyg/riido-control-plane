package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestAIAgentRunReadsIncludesCurrentlyReadableReply(t *testing.T) {
	var replyCalls atomic.Int32
	upstream := viewerReplyContextServer(t, func(w http.ResponseWriter, r *http.Request) {
		replyCalls.Add(1)
		var request struct{ Query string }
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if strings.Contains(request.Query, "body") || !strings.Contains(request.Query, "getCommentReplies") {
			t.Errorf("expected reply metadata query, got %q", request.Query)
		}
		writeTaskContextFixture(t, w, map[string]any{"data": map[string]any{"r0": []map[string]any{{"id": "reply", "status": "active", "updatedAt": "2026-10-26T00:00:00Z"}}}})
	})
	server, _, _ := runReadsTestServer(t, []AssignmentContextRead{
		{SourceKind: "comment", SourceID: "reply", SourcePart: "body", SourceVersion: assignmentContextVersion("")},
		{SourceKind: "comment", SourceID: "reply", SourcePart: "body"},
	})
	client, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL})
	if err != nil {
		t.Fatal(err)
	}
	server.taskContext = client
	rows := runReadsMetadataRows(t, server, "/v1/client/ai-agent/runs/run/reads")
	if len(rows) != 2 || replyCalls.Load() != 1 {
		t.Fatalf("readable reply excluded: rows=%d reply calls=%d", len(rows), replyCalls.Load())
	}
	for _, row := range rows {
		if _, exists := row["modified_since_read"]; exists {
			t.Fatalf("metadata-only reply cannot prove unchanged body: %v", row)
		}
	}
}

func viewerReplyContextServer(t *testing.T, replies http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer viewer-token" {
			t.Error("current viewer bearer was not forwarded")
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		switch r.URL.Path {
		case "/public/components/task/workspace":
			writeTaskContextFixture(t, w, map[string]any{"id": "task", "team": map[string]any{"id": "team", "workspace": map[string]any{"id": defaultAIAgentClientWorkspaceID}}})
		case "/teams/team/components/task":
			writeTaskContextFixture(t, w, map[string]any{"id": "task", "componentType": "task", "document": map[string]any{"HTMLContent": "<p>current</p>"}})
		case "/documents/providers/team/task":
			writeTaskContextFixture(t, w, map[string]any{"html": "<p>current</p>"})
		case "/teams/team/components/task/comments":
			writeTaskContextFixture(t, w, map[string]any{"comments": []map[string]any{{"id": "parent", "status": "active", "replyCount": 1}}, "meta": map[string]any{"total": 1}})
		case "/teams/team/components/task/web-links":
			writeTaskContextFixture(t, w, map[string]any{"links": []any{}})
		case "/teams/team/components/task/files", "/teams/team/components/task/relations/issues", "/teams/team/components/task/relations/meetings":
			writeTaskContextFixture(t, w, []any{})
		case "/graphql":
			replies(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server
}
