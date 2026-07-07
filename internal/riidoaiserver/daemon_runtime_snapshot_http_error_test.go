package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPDaemonRuntimeSnapshotRejectsInvalidBoundaries(t *testing.T) {
	unconfigured := NewServer(ServerConfig{}).Handler()
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/daemon/runtime-snapshot", nil)
	unconfigured.ServeHTTP(resp, req)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("method status=%d body=%s", resp.Code, resp.Body.String())
	}
	resp = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v1/daemon/runtime-snapshot", nil)
	unconfigured.ServeHTTP(resp, req)
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("unconfigured status=%d body=%s", resp.Code, resp.Body.String())
	}

	server := newRuntimeSnapshotHTTPTestServer(t, NewDevelopmentAIAgentClientStore())
	for _, tc := range []struct {
		name     string
		deviceID string
		secret   string
		want     int
	}{
		{name: "missing device id", want: http.StatusUnauthorized},
		{name: "missing secret", deviceID: "device-a", want: http.StatusUnauthorized},
		{name: "bad secret", deviceID: "device-a", secret: "bad", want: http.StatusUnauthorized},
	} {
		req := httptest.NewRequest(http.MethodPost, "/v1/daemon/runtime-snapshot", strings.NewReader(`{}`))
		req.Header.Set(deviceIDHeader, tc.deviceID)
		req.Header.Set(deviceSecretHeader, tc.secret)
		resp := httptest.NewRecorder()
		server.ServeHTTP(resp, req)
		if resp.Code != tc.want {
			t.Fatalf("%s status=%d body=%s", tc.name, resp.Code, resp.Body.String())
		}
	}
}
