package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func assertTaskCommentStopReplayEvents(t *testing.T, server http.Handler) {
	t.Helper()
	eventsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/events?replay=1", nil)
	eventsReq.Header.Set("Authorization", "Bearer user-token")
	eventsResp := httptest.NewRecorder()
	server.ServeHTTP(eventsResp, eventsReq)
	body := eventsResp.Body.String()
	if !strings.Contains(body, string(AgentTaskCommentQueuedByBusyAgent)) ||
		!strings.Contains(body, string(AgentTaskCommentStoppedByUserRequest)) {
		t.Fatalf("events body = %q", body)
	}
}
