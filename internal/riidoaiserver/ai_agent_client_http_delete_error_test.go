package riidoaiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAIAgentClientDeleteErrors(t *testing.T) {
	t.Run("missing auth", func(t *testing.T) {
		server := newDeleteErrorTestServer(t, NewDevelopmentAIAgentClientStore(), nil)
		assertDeleteStatus(t, server, "", http.StatusUnauthorized)
	})
	t.Run("active thread read error", func(t *testing.T) {
		assignment := NewStore()
		t.Cleanup(assignment.Close)
		server := newDeleteErrorTestServer(t, deleteActiveThreadsErrorStore{
			DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(),
			err:                           errors.New("active thread read failed"),
		}, assignment)
		assertDeleteStatus(t, server, "user-token", http.StatusBadRequest)
	})
	t.Run("cancel error", func(t *testing.T) {
		assignment := deleteCancelErrorStore{Store: NewStore(), err: errors.New("cancel failed")}
		t.Cleanup(assignment.Close)
		server := newDeleteErrorTestServer(t, deleteActiveThreadsStore{
			DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(),
		}, assignment)
		assertDeleteStatus(t, server, "user-token", http.StatusBadRequest)
	})
	t.Run("delete error", func(t *testing.T) {
		server := newDeleteErrorTestServer(t, deleteAgentErrorStore{
			DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(),
			err:                           errors.New("delete failed"),
		}, nil)
		assertDeleteStatus(t, server, "user-token", http.StatusBadRequest)
	})
}

func assertDeleteStatus(t *testing.T, server http.Handler, token string, want int) {
	t.Helper()
	req := httptest.NewRequest(http.MethodDelete, "/v1/client/ai-agent/agents/agent-owned-codex", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != want {
		t.Fatalf("delete status=%d want=%d body=%s", resp.Code, want, resp.Body.String())
	}
}
