package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newRuntimeSnapshotHTTPTestServer(t *testing.T, store *DevelopmentAIAgentClientStore) http.Handler {
	t.Helper()
	auth, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-runtime-snapshot",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	return NewServer(ServerConfig{AIAgentClient: store, Authorizer: auth}).Handler()
}

func enrollRuntimeSnapshotDevice(t *testing.T, server http.Handler) EnrollDeviceResponse {
	t.Helper()
	body := `{"display_name":"Runtime Snapshot Mac","platform":"darwin"}`
	req := httptest.NewRequest(http.MethodPost,
		"/v2/desktop/workspaces/workspace-alpha/devices/enroll",
		strings.NewReader(body))
	req.Header.Set(aiAgentTokenHeader, "user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("enroll status=%d body=%s", resp.Code, resp.Body.String())
	}
	var enrollment EnrollDeviceResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &enrollment); err != nil {
		t.Fatalf("enroll json: %v", err)
	}
	return enrollment
}

func runtimeSnapshotRequest(body string, enrollment EnrollDeviceResponse) *http.Request {
	req := httptest.NewRequest(http.MethodPost,
		"/v1/daemon/runtime-snapshot",
		strings.NewReader(body))
	req.Header.Set(deviceIDHeader, enrollment.DeviceID)
	req.Header.Set(deviceSecretHeader, enrollment.DeviceSecret)
	return req
}
