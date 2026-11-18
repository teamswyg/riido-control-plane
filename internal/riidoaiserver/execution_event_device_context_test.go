package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// An already authenticated device must not depend on the mutable provider
// availability registry merely to report an admitted execution's progress.
func TestOrdinaryAgentEventKeepsAuthenticatedDevice(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(deviceIDHeader, "device-a")
	w := httptest.NewRecorder()
	ctx, ok := authenticatedCoverageEventContext(w, req, AgentEventRequest{DeviceID: "device-a"}, AuthorizationResult{PrincipalID: "owner"})
	if !ok {
		t.Fatalf("status=%d", w.Code)
	}
	got, ok := claimDeviceFromContext(ctx)
	if !ok || got.deviceID != "device-a" || got.ownerPrincipalID != "owner" {
		t.Fatal("authenticated device lost; progress falls back to unavailable runtime binding")
	}
}

func TestOrdinaryAgentEventRejectsDeviceSubstitution(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(deviceIDHeader, "device-a")
	w := httptest.NewRecorder()
	if _, ok := authenticatedCoverageEventContext(w, req, AgentEventRequest{DeviceID: "device-b"}, AuthorizationResult{PrincipalID: "owner"}); ok || w.Code != http.StatusForbidden {
		t.Fatalf("device substitution accepted: %d", w.Code)
	}
}

func TestBearerEventDoesNotAcquireDeviceProvenance(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	ctx, ok := authenticatedCoverageEventContext(httptest.NewRecorder(), req, AgentEventRequest{DeviceID: "untrusted"}, AuthorizationResult{PrincipalID: "owner"})
	if !ok {
		t.Fatal("legacy bearer path rejected")
	}
	if _, ok := claimDeviceFromContext(ctx); ok {
		t.Fatal("body minted device provenance")
	}
}
