package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const approvalChatBody = "<p>너가 직접 진행해줘 승인할게</p>"

func TestHTTPThreadMessageApprovesPendingToolApproval(t *testing.T) {
	server := newApprovalChatTestServer(t)
	assigned := assignApprovalRoundTripTask(t, server)
	pollApprovalRoundTripTask(t, server, assigned.AssignmentID)
	createApprovalRoundTripRequest(t, server, assigned.AssignmentID)

	waitDone := make(chan ToolApprovalWaitResponse, 1)
	go func() {
		waitDone <- waitApprovalRoundTripDecision(t, server, assigned.AssignmentID)
	}()
	time.Sleep(50 * time.Millisecond)

	reply := postApprovalChatThreadMessage(t, server, assigned)
	if reply.AssignmentID != assigned.AssignmentID || reply.ThreadID != assigned.ThreadID {
		t.Fatalf("approval reply created new work: assigned=%+v reply=%+v", assigned, reply)
	}
	assertApprovalChatWaitApproved(t, waitDone)
	assertApprovalChatHistoryMessage(t, server, assigned.ThreadID)
}

func newApprovalChatTestServer(t *testing.T) http.Handler {
	t.Helper()
	return newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "round-trip-token",
		Scopes:      []string{"ai-agent:*", "agent:agent-public-openclaw:*"},
	}})
}

func postApprovalChatThreadMessage(
	t *testing.T,
	server http.Handler,
	assigned AIAgentTaskActionResponse,
) AIAgentTaskActionResponse {
	t.Helper()
	body := `{"body":"` + approvalChatBody + `","source_message_id":"approval-chat-1"}`
	path := "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-approval/threads/" + assigned.ThreadID + "/messages"
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer round-trip-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusAccepted {
		t.Fatalf("thread message status=%d body=%s", resp.Code, resp.Body.String())
	}
	var out AIAgentTaskActionResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("thread message json: %v", err)
	}
	return out
}
