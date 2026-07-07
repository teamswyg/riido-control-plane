package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func assignApprovalRoundTripTask(t *testing.T, server http.Handler) AIAgentTaskActionResponse {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-approval/assignment", strings.NewReader(`{"agent_id":"agent-public-openclaw"}`))
	req.Header.Set("Authorization", "Bearer round-trip-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", resp.Code, resp.Body.String())
	}
	var out AIAgentTaskActionResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("assign json: %v", err)
	}
	if out.AssignmentID == "" {
		t.Fatalf("assign response missing assignment_id: %+v", out)
	}
	return out
}

func pollApprovalRoundTripTask(t *testing.T, server http.Handler, assignmentID string) {
	t.Helper()
	body := `{"daemon_id":"daemon-shared-studio","device_id":"device-shared-studio","runtime_id":"runtime-openclaw-shared"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-public-openclaw/poll", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer round-trip-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("poll status=%d body=%s", resp.Code, resp.Body.String())
	}
	var out PollResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("poll json: %v", err)
	}
	if out.Action != PollStart || out.Assignment == nil || out.Assignment.ID != assignmentID {
		t.Fatalf("poll response = %+v", out)
	}
}
