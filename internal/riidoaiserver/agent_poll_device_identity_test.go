package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAgentPollRejectsAuthenticatedDeviceBodyMismatch(t *testing.T) {
	devices := NewDevelopmentAIAgentClientStore()
	devices.deviceCredentials["other-device"] = deviceCredentialRecord{
		deviceID: "other-device", secretHash: sha256.Sum256([]byte("fixture-secret")), ownerPrincipalID: "other-owner",
	}
	called := false
	assignments := &handlerAssignmentStore{poll: func(context.Context, string, PollRequest) (PollResponse, error) {
		called = true
		return PollResponse{Action: PollStart}, nil
	}}
	server := NewServer(ServerConfig{AIAgentClient: devices, Assignment: assignments}).Handler()
	req := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-owned-codex/poll", strings.NewReader(`{"daemon_id":"daemon-dev-macbook","device_id":"device-dev-macbook","runtime_id":"runtime-codex-dev"}`))
	req.Header.Set(deviceIDHeader, "other-device")
	req.Header.Set(deviceSecretHeader, "fixture-secret")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden || called {
		t.Fatalf("device/body mismatch: status=%d assignment store called=%v", resp.Code, called)
	}
}
