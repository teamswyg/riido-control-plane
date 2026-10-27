package riidoaiserver

import (
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestContextReadStreamRechecksReplyAccess(t *testing.T) {
	for _, mode := range []string{"live", "replay"} {
		t.Run(mode, func(t *testing.T) {
			var calls atomic.Int32
			upstream := viewerReplyContextServer(t, func(w http.ResponseWriter, r *http.Request) {
				if calls.Add(1) > 1 {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
				writeTaskContextFixture(t, w, map[string]any{"data": map[string]any{"r0": []map[string]any{{"id": "reply", "status": "active"}}}})
			})
			server, _, _, event := contextReadStreamServer(t)
			client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL})
			server.taskContext = client
			payload := event.Payload.(AgentContextReadEvent)
			payload.Sources = []AgentContextReadSource{{Kind: "task", ID: "task", Part: "body", Status: "served"}, {Kind: "comment", ID: "reply", Part: "body", Status: "served"}}
			event.Payload = payload
			second := event
			second.Seq = 2
			body := contextReadStreamRequest(server, mode, event, second).Body.String()
			if strings.Count(body, "event: agent_context_read") != 2 || strings.Count(body, `"id":"reply"`) != 1 || calls.Load() != 2 {
				t.Fatalf("reply SSE did not recheck: calls=%d body=%s", calls.Load(), body)
			}
		})
	}
}
