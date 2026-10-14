package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
	path := "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-approval/tool-approvals"
	req := httptest.NewRequest(http.MethodGet, path, nil)
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
