package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func assertNonAdminRuntimeCreateDenied(t *testing.T, createBody []byte) {
	t.Helper()
	nonAdminServer := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-2",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*"},
	}})
	deniedReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/agents", strings.NewReader(string(createBody)))
	deniedReq.Header.Set("Authorization", "Bearer user-token")
	deniedResp := httptest.NewRecorder()
	nonAdminServer.ServeHTTP(deniedResp, deniedReq)
	if deniedResp.Code != http.StatusBadRequest {
		t.Fatalf("non-admin create status=%d body=%s", deniedResp.Code, deniedResp.Body.String())
	}
}
