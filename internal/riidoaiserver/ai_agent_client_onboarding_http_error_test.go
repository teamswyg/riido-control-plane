package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type onboardingErrorStore struct {
	*DevelopmentAIAgentClientStore
	listErr   error
	createErr error
}

func (s onboardingErrorStore) ListAIAgentOnboardingFixtures(ctx context.Context, principal AuthorizationResult) (AgentOnboardingFixtureListResponse, error) {
	if s.listErr != nil {
		return AgentOnboardingFixtureListResponse{}, s.listErr
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentOnboardingFixtures(ctx, principal)
}

func (s onboardingErrorStore) CreateAIAgentFromOnboardingFixture(ctx context.Context, principal AuthorizationResult, fixtureID string, req CreateAgentConfigurationRequest) (AgentClientRecordResponse, error) {
	if s.createErr != nil {
		return AgentClientRecordResponse{}, s.createErr
	}
	return s.DevelopmentAIAgentClientStore.CreateAIAgentFromOnboardingFixture(ctx, principal, fixtureID, req)
}

func onboardingErrorTestServer(t *testing.T, listErr, createErr error) http.Handler {
	t.Helper()
	return NewServer(ServerConfig{
		AIAgentClient: onboardingErrorStore{
			DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(),
			listErr:                       listErr,
			createErr:                     createErr,
		},
		Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1"),
	}).Handler()
}

func TestHTTPAIAgentClientOnboardingErrorBranches(t *testing.T) {
	tests := []struct {
		name, method, path, body, wantMsg string
		server                            http.Handler
		want                              int
	}{
		{"missing store", http.MethodGet, "/v1/client/ai-agent/onboarding/fixtures", "", "ai agent client store is not configured", NewServer(ServerConfig{}).Handler(), http.StatusServiceUnavailable},
		{"wrong method", http.MethodPost, "/v1/client/ai-agent/onboarding/fixtures", "", "method not allowed", onboardingErrorTestServer(t, nil, nil), http.StatusMethodNotAllowed},
		{"list error", http.MethodGet, "/v1/client/ai-agent/onboarding/fixtures", "", "fixture list failed", onboardingErrorTestServer(t, errors.New("fixture list failed"), nil), http.StatusBadRequest},
		{"bad create json", http.MethodPost, "/v1/client/ai-agent/onboarding/fixtures/riido_pm/agents", `{"name":`, "unexpected EOF", onboardingErrorTestServer(t, nil, nil), http.StatusBadRequest},
		{"create error", http.MethodPost, "/v1/client/ai-agent/onboarding/fixtures/riido_pm/agents", `{"name":"A"}`, "fixture create failed", onboardingErrorTestServer(t, nil, errors.New("fixture create failed")), http.StatusBadRequest},
		{"unknown suffix", http.MethodPost, "/v1/client/ai-agent/onboarding/fixtures/riido_pm/unknown", "", "not found", onboardingErrorTestServer(t, nil, nil), http.StatusNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set(aiAgentTokenHeader, "ai-agent-token")
			resp := httptest.NewRecorder()
			tc.server.ServeHTTP(resp, req)
			if resp.Code != tc.want || !strings.Contains(resp.Body.String(), tc.wantMsg) {
				t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
			}
		})
	}
}
