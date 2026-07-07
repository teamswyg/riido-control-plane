package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type failingDeviceDaemonControlStore struct {
	*DevelopmentAIAgentClientStore
	err error
}

func (s *failingDeviceDaemonControlStore) ControlAIAgentDeviceDaemon(
	context.Context,
	AuthorizationResult,
	string,
	DaemonControlAction,
	ControlDeviceDaemonRequest,
) (DeviceDaemonCommandResponse, error) {
	return DeviceDaemonCommandResponse{}, s.err
}

func TestHTTPAIAgentClientDeviceDaemonControlErrorBranches(t *testing.T) {
	t.Run("invalid action", func(t *testing.T) {
		server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
			PrincipalID: "user-1",
			Token:       "user-token",
			Scopes:      []string{"ai-agent:*"},
		}})
		req := httptest.NewRequest(http.MethodPost, "/v2/client/workspaces/workspace-alpha/ai-agent/devices/device-dev-macbook/daemon/reload", nil)
		req.Header.Set("Authorization", "Bearer user-token")
		resp := httptest.NewRecorder()
		server.ServeHTTP(resp, req)
		if resp.Code != http.StatusNotFound {
			t.Fatalf("invalid action status=%d body=%s", resp.Code, resp.Body.String())
		}
	})

	t.Run("malformed body", func(t *testing.T) {
		server := newDeviceDaemonsErrorTestServer(t, []string{"ai-agent:*"}, NewDevelopmentAIAgentClientStore())
		req := httptest.NewRequest(http.MethodPost, "/v2/client/workspaces/workspace-alpha/ai-agent/devices/device-dev-macbook/daemon/restart", strings.NewReader(`{`))
		req.Header.Set("Authorization", "Bearer user-token")
		resp := httptest.NewRecorder()
		server.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest {
			t.Fatalf("malformed body status=%d body=%s", resp.Code, resp.Body.String())
		}
	})

	t.Run("backend error", func(t *testing.T) {
		store := &failingDeviceDaemonControlStore{DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(), err: errors.New("device control failed")}
		server := newDeviceDaemonsErrorTestServer(t, []string{"ai-agent:*"}, store)
		req := httptest.NewRequest(http.MethodPost, "/v2/client/workspaces/workspace-alpha/ai-agent/devices/device-dev-macbook/daemon/restart", strings.NewReader(`{"reason":"qa"}`))
		req.Header.Set("Authorization", "Bearer user-token")
		resp := httptest.NewRecorder()
		server.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest {
			t.Fatalf("backend error status=%d body=%s", resp.Code, resp.Body.String())
		}
	})
}
