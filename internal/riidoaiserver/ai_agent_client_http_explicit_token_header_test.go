package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAIAgentClientDevelopmentAcceptsExplicitAIAgentTokenHeader(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:read"},
	}})

	req := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/bootstrap", nil)
	req.Header.Set(aiAgentTokenHeader, "user-token")
	req.Header.Set("Authorization", "Bearer wrong-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", resp.Code, resp.Body.String())
	}
}
