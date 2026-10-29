package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCoveredBatchProgressRetainsAuthenticatedGeneration(t *testing.T) {
	devices := NewDevelopmentAIAgentClientStore()
	devices.deviceCredentials["dcov_fixture"] = deviceCredentialRecord{deviceID: "dcov_fixture", ownerPrincipalID: "owner", secretHash: sha256.Sum256([]byte("fixture-secret")), coverageGeneration: "coverage-fixture", credentialGeneration: "credential-fixture"}
	called := 0
	store := &handlerAssignmentStore{record: func(ctx context.Context, _ string, _ AgentEventRequest) (AgentEventResponse, error) {
		auth, ok := ctx.Value(executionCoverageAuthKey{}).(executionCoverageAuth)
		if !ok || auth.generation != "coverage-fixture" || auth.credential != "credential-fixture" {
			return AgentEventResponse{}, ErrAuthorizationForbidden
		}
		called++
		return AgentEventResponse{SchemaVersion: SchemaVersion}, nil
	}}
	server := Server{assignment: store, devices: devices}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(progressBody(`"device_id":"dcov_fixture","runtime_id":"runtime-a",`)))
	req.Header.Set(deviceIDHeader, "dcov_fixture")
	req.Header.Set(deviceSecretHeader, "fixture-secret")
	req.Header.Set(executionCoverageHeader, "coverage-fixture")
	w := httptest.NewRecorder()
	server.handleAgentThreadProgress(w, req, "agent-a")
	if w.Code != http.StatusAccepted || called != 2 {
		t.Fatalf("covered progress lost auth context: status=%d calls=%d", w.Code, called)
	}
}
