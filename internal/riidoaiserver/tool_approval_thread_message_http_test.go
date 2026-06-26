package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
	chatBody string,
) AIAgentTaskActionResponse {
	t.Helper()
	body := `{"body":"` + chatBody + `","source_message_id":"approval-chat-1"}`
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
