package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCurrentReplyMetadataFailsClosed(t *testing.T) {
	for _, mode := range []string{"forbidden", "partial-error", "malformed", "oversize", "missing-alias", "partial-alias", "null-alias"} {
		t.Run(mode, func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch mode {
				case "forbidden":
					http.Error(w, "forbidden", http.StatusForbidden)
				case "partial-error":
					writeTaskContextFixture(t, w, map[string]any{"data": map[string]any{"r0": []map[string]any{{"id": "reply", "status": "active"}}}, "errors": []map[string]any{{"extensions": map[string]any{"code": "FORBIDDEN"}}}})
				case "malformed":
					_, _ = w.Write([]byte("not json"))
				case "oversize":
					writeTaskContextFixture(t, w, map[string]any{"data": map[string]any{"r0": []map[string]any{{"id": strings.Repeat("x", 1<<20), "status": "active"}}}})
				case "missing-alias":
					writeTaskContextFixture(t, w, map[string]any{"data": map[string]any{"r1": []map[string]any{{"id": "reply", "status": "active"}}}})
				case "partial-alias":
					writeTaskContextFixture(t, w, map[string]any{"data": map[string]any{"r0": []map[string]any{{"id": "reply", "status": "active"}}}})
				case "null-alias":
					writeTaskContextFixture(t, w, map[string]any{"data": map[string]any{"r0": []map[string]any{{"id": "reply", "status": "active"}}, "r1": nil}})
				}
			}))
			t.Cleanup(upstream.Close)
			client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL})
			if got := client.readCurrentReplyMetadata(t.Context(), "team", "viewer", []AIAgentTaskContextComment{{ID: "parent", ReplyCount: 1}, {ID: "parent-2", ReplyCount: 1}}); len(got) != 0 {
				t.Fatalf("reply exposed after %s", mode)
			}
		})
	}
}
