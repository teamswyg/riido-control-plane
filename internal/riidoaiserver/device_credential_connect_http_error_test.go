package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type credentialConnectErrorStore struct {
	*DevelopmentAIAgentClientStore
	err error
}

func (s *credentialConnectErrorStore) ConnectAIAgentDeviceWithCredential(context.Context, AuthorizationResult, string, string) (DeviceRecord, error) {
	return DeviceRecord{}, s.err
}

func TestDesktopCredentialConnectSanitizesDependencyErrors(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
	}{
		{errors.New("private storage detail"), http.StatusServiceUnavailable},
		{fmt.Errorf("private storage detail: %w", context.DeadlineExceeded), http.StatusGatewayTimeout},
		{ErrAuthorizationUnauthenticated, http.StatusUnauthorized},
		{ErrAuthorizationForbidden, http.StatusForbidden},
	} {
		store := &credentialConnectErrorStore{NewDevelopmentAIAgentClientStore(), tc.err}
		handler := NewServer(ServerConfig{AIAgentClient: store, Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "guest")}).Handler()
		req := httptest.NewRequest(http.MethodPost, "/v2/desktop/workspaces/workspace/devices/connect", strings.NewReader(`{"device_id":"device"}`))
		req.Header.Set(aiAgentTokenHeader, "ai-agent-token")
		req.Header.Set(deviceSecretHeader, "synthetic-device-proof")
		out := httptest.NewRecorder()
		handler.ServeHTTP(out, req)
		if out.Code != tc.status {
			t.Fatalf("status=%d want=%d", out.Code, tc.status)
		}
		if strings.Contains(out.Body.String(), "private storage detail") {
			t.Fatal("dependency detail leaked")
		}
	}
}
