package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientDevelopmentSSEHidesStaleQueuedStatus(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:stream"},
	}})

	req := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/events?replay=1", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("events status=%d body=%s", resp.Code, resp.Body.String())
	}
	if got := resp.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
		t.Fatalf("Content-Type = %q", got)
	}
	body := resp.Body.String()
	if strings.Contains(body, string(AgentTaskCommentQueuedByBusyAgent)) {
		t.Fatalf("stale queued status leaked into replay: %q", body)
	}
}

func TestHTTPAIAgentClientDevelopmentSSEHidesCurrentQueuedStatus(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*"},
	}})

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/client/ai-agent/tasks/task-current-queued/comments",
		strings.NewReader(`{"agent_id":"agent-owned-codex","body":"queued please"}`),
	)
	req.Header.Set("Authorization", "Bearer user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusAccepted {
		t.Fatalf("comment status=%d body=%s", resp.Code, resp.Body.String())
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/events?replay=1", nil)
	eventsReq.Header.Set("Authorization", "Bearer user-token")
	eventsResp := httptest.NewRecorder()
	server.ServeHTTP(eventsResp, eventsReq)
	body := eventsResp.Body.String()
	if strings.Contains(body, string(AgentTaskCommentQueuedByBusyAgent)) {
		t.Fatalf("current queued status leaked into replay: %q", body)
	}
}
