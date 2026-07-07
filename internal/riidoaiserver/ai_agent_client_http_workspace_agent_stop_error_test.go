package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientWorkspaceAgentStopErrors(t *testing.T) {
	const token = "user-token"
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       token,
		Scopes:      []string{"ai-agent:*"},
	}})

	for _, tc := range []struct {
		name   string
		method string
		path   string
		body   string
		token  string
		want   int
	}{
		{
			name:   "method rejected as route miss",
			method: http.MethodGet,
			path:   base + "/agent-assignments/agent-public-openclaw/stop",
			want:   http.StatusNotFound,
		},
		{
			name:   "malformed suffix",
			method: http.MethodPost,
			path:   base + "/agent-assignments/agent-public-openclaw/extra/stop",
			token:  token,
			want:   http.StatusNotFound,
		},
		{
			name:   "missing auth",
			method: http.MethodPost,
			path:   base + "/agent-assignments/agent-public-openclaw/stop",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "bad json",
			method: http.MethodPost,
			path:   base + "/agent-assignments/agent-public-openclaw/stop",
			body:   `{`,
			token:  token,
			want:   http.StatusBadRequest,
		},
		{
			name:   "no active threads",
			method: http.MethodPost,
			path:   base + "/agent-assignments/agent-missing/stop",
			token:  token,
			want:   http.StatusNotFound,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			if resp.Code != tc.want {
				t.Fatalf("stop status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
		})
	}
}
