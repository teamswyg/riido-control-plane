package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const onboardingAgentCreatePath = "/v2/client/workspaces/workspace-dev-riid/ai-agent/onboarding/fixtures/riido_pm/agents"

func onboardingAgentCreateHTTPHandler(t *testing.T, store AIAgentClientStore) http.Handler {
	t.Helper()
	return NewServer(ServerConfig{AIAgentClient: store, Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1")}).Handler()
}

func onboardingAgentCreateHTTPBody(t *testing.T, requestID string) string {
	t.Helper()
	body := map[string]any{"name": "response-loss fixture", "visibility": "private", "runtime_id": "runtime-cursor-dev", "model_id": "cursor-auto"}
	if requestID != "" {
		body["request_id"] = requestID
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func postOnboardingAgentCreateHTTP(handler http.Handler, path, body, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set(aiAgentTokenHeader, token)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func decodeOnboardingAgentCreateHTTP(t *testing.T, rec *httptest.ResponseRecorder) AgentClientRecordResponse {
	t.Helper()
	if rec.Code != http.StatusCreated {
		t.Fatalf("fixture create status=%d body=%s", rec.Code, rec.Body.String())
	}
	var response AgentClientRecordResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	return response
}
