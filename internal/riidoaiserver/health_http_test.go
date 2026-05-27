package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPHealthAndReadyArePublic(t *testing.T) {
	server := NewServer(ServerConfig{}).Handler()
	for _, tc := range []struct {
		name       string
		path       string
		wantStatus string
	}{
		{name: "health", path: "/healthz", wantStatus: "ok"},
		{name: "ready", path: "/readyz", wantStatus: "ready"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			if resp.Code != http.StatusOK {
				t.Fatalf("%s status=%d body=%s", tc.path, resp.Code, resp.Body.String())
			}
			var health Health
			if err := json.Unmarshal(resp.Body.Bytes(), &health); err != nil {
				t.Fatalf("health json: %v", err)
			}
			if health.SchemaVersion != SchemaVersion || health.Status != tc.wantStatus {
				t.Fatalf("health = %+v", health)
			}
		})
	}
}

func TestHTTPHealthAndReadyRejectWrongMethod(t *testing.T) {
	server := NewServer(ServerConfig{}).Handler()
	for _, path := range []string{"/healthz", "/readyz"} {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		resp := httptest.NewRecorder()
		server.ServeHTTP(resp, req)
		if resp.Code != http.StatusMethodNotAllowed {
			t.Fatalf("%s wrong method status=%d body=%s", path, resp.Code, resp.Body.String())
		}
	}
}
