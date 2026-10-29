package riidoaiserver

import (
	"net/http"
	"sync/atomic"
	"testing"
)

func TestAIAgentRunReadsRechecksReplyAccess(t *testing.T) {
	for _, mode := range []string{"deleted", "forbidden", "partial-error"} {
		t.Run(mode, func(t *testing.T) {
			var calls atomic.Int32
			upstream := viewerReplyContextServer(t, func(w http.ResponseWriter, r *http.Request) {
				if calls.Add(1) == 1 {
					writeTaskContextFixture(t, w, map[string]any{"data": map[string]any{"r0": []map[string]any{{"id": "reply", "status": "active"}}}})
					return
				}
				switch mode {
				case "deleted":
					writeTaskContextFixture(t, w, map[string]any{"data": map[string]any{"r0": []map[string]any{{"id": "reply", "status": "deleted"}}}})
				case "forbidden":
					http.Error(w, "forbidden", http.StatusForbidden)
				case "partial-error":
					writeTaskContextFixture(t, w, map[string]any{"data": map[string]any{"r0": []map[string]any{{"id": "reply", "status": "active"}}}, "errors": []map[string]any{{"extensions": map[string]any{"code": "FORBIDDEN"}}}})
				}
			})
			server, _, _ := runReadsTestServer(t, []AssignmentContextRead{{SourceKind: "comment", SourceID: "parent", SourcePart: "body"}, {SourceKind: "comment", SourceID: "reply", SourcePart: "body"}})
			client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL})
			server.taskContext = client
			first := runReadsMetadataRows(t, server, "/v1/client/ai-agent/runs/run/reads")
			second := runReadsMetadataRows(t, server, "/v1/client/ai-agent/runs/run/reads")
			if len(first) != 2 || len(second) != 1 || second[0]["id"] != "parent" || calls.Load() != 2 {
				t.Fatalf("stale reply visibility: first=%d second=%v calls=%d", len(first), second, calls.Load())
			}
		})
	}
}
