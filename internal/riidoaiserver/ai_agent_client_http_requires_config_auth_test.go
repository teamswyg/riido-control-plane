package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientDevelopmentRequiresConfigurationAndAuthorization(t *testing.T) {
	unconfigured := NewServer(ServerConfig{Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1")}).Handler()
	req := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/bootstrap", nil)
	req.Header.Set("Authorization", "Bearer ai-agent-token")
	resp := httptest.NewRecorder()
	unconfigured.ServeHTTP(resp, req)
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("unconfigured status=%d body=%s", resp.Code, resp.Body.String())
	}

	configured := NewServer(ServerConfig{AIAgentClient: NewDevelopmentAIAgentClientStore(), Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:stream"}, "user-1")}).Handler()
	forbiddenReq := httptest.NewRequest(http.MethodPatch, "/v1/client/ai-agent/agents/agent-owned-claude", strings.NewReader(`{"name":"nope"}`))
	forbiddenReq.Header.Set("Authorization", "Bearer ai-agent-token")
	forbiddenResp := httptest.NewRecorder()
	configured.ServeHTTP(forbiddenResp, forbiddenReq)
	if forbiddenResp.Code != http.StatusForbidden {
		t.Fatalf("forbidden patch status=%d body=%s", forbiddenResp.Code, forbiddenResp.Body.String())
	}
}
