package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func waitApprovalRoundTripDecision(t *testing.T, server http.Handler, assignmentID string) ToolApprovalWaitResponse {
	t.Helper()
	body := `{"assignment_id":"` + assignmentID + `","wait_ms":5000}`
	req := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-public-openclaw/tool-approvals/approval-1/wait", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer round-trip-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("wait approval status=%d body=%s", resp.Code, resp.Body.String())
	}
	var out ToolApprovalWaitResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("wait approval json: %v", err)
	}
	return out
}

func decideApprovalRoundTrip(t *testing.T, server http.Handler, assignmentID string) ToolApprovalDecisionResponse {
	t.Helper()
	body := `{"assignment_id":"` + assignmentID + `","decision":"approve","reason":"approved from web"}`
	path := "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-approval/tool-approvals/approval-1/decision"
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer round-trip-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("decide approval status=%d body=%s", resp.Code, resp.Body.String())
	}
	var out ToolApprovalDecisionResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("decide approval json: %v", err)
	}
	return out
}
