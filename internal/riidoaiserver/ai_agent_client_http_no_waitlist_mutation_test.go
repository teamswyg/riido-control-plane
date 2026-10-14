package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientDevelopmentDoesNotExposeWaitlistMarketingMutation(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*"},
	}})

	paths := []string{
		"/v1/client/ai-agent/waitlist",
		"/v1/client/ai-agent/runtime-waitlist",
		"/v1/client/ai-agent/marketing-consent",
	}
	for _, path := range paths {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"email":"viewer@example.com"}`))
		req.Header.Set("Authorization", "Bearer user-token")
		resp := httptest.NewRecorder()
		server.ServeHTTP(resp, req)
		if resp.Code != http.StatusNotFound {
			t.Fatalf("%s should not be exposed by AI Agent client API, status=%d body=%s", path, resp.Code, resp.Body.String())
		}
	}
}
