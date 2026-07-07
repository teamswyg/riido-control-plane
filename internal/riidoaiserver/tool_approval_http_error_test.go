package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPToolApprovalRoutesRequireStore(t *testing.T) {
	server := NewServer(ServerConfig{
		Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1"),
	}).Handler()
	for _, tc := range []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"create", http.MethodPost, "/v1/agents/agent-public-openclaw/tool-approvals", `{}`},
		{"wait", http.MethodPost, "/v1/agents/agent-public-openclaw/tool-approvals/apr-1/wait", `{}`},
		{"list", http.MethodGet, "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-a/tool-approvals", ""},
		{"decide", http.MethodPost, "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-a/tool-approvals/apr-1/decision", `{}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertToolApprovalHTTPStatus(t, server, tc.method, tc.path, tc.body, http.StatusServiceUnavailable)
		})
	}
}

func TestHTTPToolApprovalRoutesRejectMalformedJSON(t *testing.T) {
	server := newApprovalChatTestServer(t)
	for _, tc := range []struct {
		name   string
		method string
		path   string
	}{
		{"create", http.MethodPost, "/v1/agents/agent-public-openclaw/tool-approvals"},
		{"wait", http.MethodPost, "/v1/agents/agent-public-openclaw/tool-approvals/apr-1/wait"},
		{"decide", http.MethodPost, "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-a/tool-approvals/apr-1/decision"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertToolApprovalHTTPStatus(t, server, tc.method, tc.path, `{"broken":`, http.StatusBadRequest)
		})
	}
}

func assertToolApprovalHTTPStatus(
	t *testing.T,
	server http.Handler,
	method string,
	path string,
	body string,
	want int,
) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer round-trip-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != want {
		t.Fatalf("%s %s status=%d want=%d body=%s", method, path, resp.Code, want, resp.Body.String())
	}
}
