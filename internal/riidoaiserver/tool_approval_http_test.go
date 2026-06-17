package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPToolApprovalRoundTrip(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "round-trip-token",
		Scopes: []string{
			"ai-agent:*",
			"agent:agent-public-openclaw:*",
		},
	}})
	assigned := assignApprovalRoundTripTask(t, server)
	pollApprovalRoundTripTask(t, server, assigned.AssignmentID)
	createApprovalRoundTripRequest(t, server, assigned.AssignmentID)
	assertApprovalRoundTripList(t, server, assigned.AssignmentID)

	waitDone := make(chan ToolApprovalWaitResponse, 1)
	go func() {
		waitDone <- waitApprovalRoundTripDecision(t, server, assigned.AssignmentID)
	}()
	time.Sleep(50 * time.Millisecond)
	decision := decideApprovalRoundTrip(t, server, assigned.AssignmentID)
	if decision.Result.Status != ApprovalApproved || decision.Decision == nil {
		t.Fatalf("decision response = %+v", decision)
	}
	select {
	case waited := <-waitDone:
		if waited.Result.Status != ApprovalApproved || waited.Decision == nil ||
			waited.Decision.DecidedBy != "user-1" {
			t.Fatalf("wait response = %+v", waited)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for approval decision")
	}
}

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
	req := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-public-openclaw/poll", strings.NewReader(`{"daemon_id":"daemon-shared-studio","device_id":"device-shared-studio","runtime_id":"runtime-openclaw-shared"}`))
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

func createApprovalRoundTripRequest(t *testing.T, server http.Handler, assignmentID string) {
	t.Helper()
	body := `{"approval_id":"approval-1","assignment_id":"` + assignmentID + `","task_id":"task-approval","tool_id":"tool-1","tool_kind":"patch_apply","tool_name":"apply_patch","reason":"provider requested tool approval"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-public-openclaw/tool-approvals", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer round-trip-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create approval status=%d body=%s", resp.Code, resp.Body.String())
	}
	var out ToolApprovalCreateResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("create approval json: %v", err)
	}
	if out.Approval.AgentID != "agent-public-openclaw" || out.Approval.Status != ApprovalPending {
		t.Fatalf("create approval response = %+v", out)
	}
}

func assertApprovalRoundTripList(t *testing.T, server http.Handler, assignmentID string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-approval/tool-approvals", nil)
	req.Header.Set("Authorization", "Bearer round-trip-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("list approvals status=%d body=%s", resp.Code, resp.Body.String())
	}
	var out ToolApprovalListResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("list approvals json: %v", err)
	}
	if len(out.Approvals) != 1 || out.Approvals[0].AssignmentID != assignmentID {
		t.Fatalf("list approvals response = %+v", out)
	}
}

func waitApprovalRoundTripDecision(t *testing.T, server http.Handler, assignmentID string) ToolApprovalWaitResponse {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-public-openclaw/tool-approvals/approval-1/wait", strings.NewReader(`{"assignment_id":"`+assignmentID+`","wait_ms":5000}`))
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
	req := httptest.NewRequest(http.MethodPost, "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-approval/tool-approvals/approval-1/decision", strings.NewReader(body))
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
