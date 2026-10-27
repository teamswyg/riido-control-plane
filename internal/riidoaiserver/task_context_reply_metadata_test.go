package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestCurrentReplyMetadataBatchesBoundedCurrentRoots(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var request struct {
			Query     string
			Variables map[string]string
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if r.Header.Get("Authorization") != "Bearer viewer" || request.Variables["team"] != "team" {
			t.Error("viewer/team scope not forwarded")
		}
		if strings.Count(request.Query, "getCommentReplies(") != 200 || strings.Contains(request.Query, "root-200") || strings.Contains(request.Query, "body") {
			t.Error("reply query must be bounded metadata aliases")
		}
		data := map[string]any{"r0": []map[string]any{{"id": "reply-0", "status": "active"}, {"id": "deleted", "status": "deleted"}, {"id": "unknown", "status": "unknown"}, {"id": "root-0", "status": "active"}}, "unexpected": []map[string]any{{"id": "injected", "status": "active"}}}
		for i := 1; i < 200; i++ {
			data[fmt.Sprintf("r%d", i)] = []map[string]any{{"id": fmt.Sprintf("reply-%d", i), "status": "active"}}
		}
		writeTaskContextFixture(t, w, map[string]any{"data": data})
	}))
	t.Cleanup(upstream.Close)
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL})
	roots := []AIAgentTaskContextComment{{ID: "no-replies"}, {ID: "not-a-root", ParentID: "other", ReplyCount: 1}}
	for i := 0; i < 201; i++ {
		roots = append(roots, AIAgentTaskContextComment{ID: fmt.Sprintf("root-%d", i), ReplyCount: 1})
	}
	roots = append(roots, roots[2])
	replies := client.readCurrentReplyMetadata(t.Context(), "team", "viewer", roots)
	if calls.Load() != 1 || len(replies) != 200 {
		t.Fatalf("calls=%d replies=%d", calls.Load(), len(replies))
	}
	for i, reply := range replies {
		if reply.ID != fmt.Sprintf("reply-%d", i) || reply.ParentID != fmt.Sprintf("root-%d", i) || !reply.MetadataOnly || !reply.Truncated || reply.Body != "" {
			t.Fatalf("reply metadata scope mismatch at %d", i)
		}
	}
	for _, source := range assignmentContextCommentSources(replies) {
		if source.Version != "" {
			t.Fatal("metadata-only replies must have no body hash")
		}
	}
	client.readCurrentReplyMetadata(t.Context(), "team", "viewer", nil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	client.readCurrentReplyMetadata(ctx, "team", "viewer", roots)
	if calls.Load() != 1 {
		t.Fatal("empty or cancelled lookups made an upstream request")
	}
}
