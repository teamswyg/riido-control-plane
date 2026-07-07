package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestHTTPToolApprovalRoutesRejectUnauthorizedScope(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "round-trip-token",
		Scopes:      []string{"component-task:unrelated:read"},
	}})
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
			assertToolApprovalHTTPStatus(t, server, tc.method, tc.path, tc.body, http.StatusForbidden)
		})
	}
}

func TestToolApprovalNestedSuffixIDRejectsEmptyID(t *testing.T) {
	for _, suffix := range []string{
		"tool-approvals//decision",
		"tool-approvals///decision",
		"tool-approvals//wait",
	} {
		t.Run(suffix, func(t *testing.T) {
			if id, ok := toolApprovalNestedSuffixID(suffix, "tool-approvals/", "/decision"); ok || id != "" {
				t.Fatalf("decision suffix id=%q ok=%v, want empty false", id, ok)
			}
		})
	}
}
