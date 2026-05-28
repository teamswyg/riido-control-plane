package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPWebFrontendCORSAllowsConfiguredOriginPreflight(t *testing.T) {
	server := NewServer(ServerConfig{WebAllowedOrigins: []string{"https://console.riido.io"}}).Handler()
	req := httptest.NewRequest(http.MethodOptions, "/v1/agent-catalog", nil)
	req.Header.Set("Origin", "https://console.riido.io")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	req.Header.Set("Access-Control-Request-Headers", "Authorization, Content-Type")
	resp := httptest.NewRecorder()

	server.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("preflight status=%d body=%s", resp.Code, resp.Body.String())
	}
	if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "https://console.riido.io" {
		t.Fatalf("allow origin = %q", got)
	}
	if got := resp.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, http.MethodPatch) || !strings.Contains(got, http.MethodDelete) {
		t.Fatalf("allow methods = %q", got)
	}
	if got := resp.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "Authorization") || !strings.Contains(got, "Content-Type") {
		t.Fatalf("allow headers = %q", got)
	}
	if got := strings.Join(resp.Header().Values("Vary"), ","); !strings.Contains(got, "Origin") {
		t.Fatalf("vary = %q", got)
	}
}

func TestHTTPWebFrontendCORSAddsHeadersToAllowedActualRequest(t *testing.T) {
	server := NewServer(ServerConfig{WebAllowedOrigins: []string{"http://localhost:5173"}}).Handler()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	resp := httptest.NewRecorder()

	server.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("health status=%d body=%s", resp.Code, resp.Body.String())
	}
	if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("allow origin = %q", got)
	}
	if got := resp.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("credentials header = %q", got)
	}
}

func TestHTTPWebFrontendCORSRejectsUnconfiguredOriginPreflight(t *testing.T) {
	server := NewServer(ServerConfig{WebAllowedOrigins: []string{"https://console.riido.io"}}).Handler()
	req := httptest.NewRequest(http.MethodOptions, "/v1/agent-catalog", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	resp := httptest.NewRecorder()

	server.ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Fatalf("preflight status=%d body=%s", resp.Code, resp.Body.String())
	}
	if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unexpected allow origin = %q", got)
	}
}

func TestHTTPWebFrontendCORSIsDisabledWithoutAllowedOrigins(t *testing.T) {
	server := NewServer(ServerConfig{}).Handler()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	resp := httptest.NewRecorder()

	server.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("health status=%d body=%s", resp.Code, resp.Body.String())
	}
	if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unexpected allow origin = %q", got)
	}
}

func TestHTTPWebFrontendCORSRejectsUnsupportedPreflightHeader(t *testing.T) {
	server := NewServer(ServerConfig{WebAllowedOrigins: []string{"https://console.riido.io"}}).Handler()
	req := httptest.NewRequest(http.MethodOptions, "/v1/agent-catalog", nil)
	req.Header.Set("Origin", "https://console.riido.io")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	req.Header.Set("Access-Control-Request-Headers", "X-Private-Debug")
	resp := httptest.NewRecorder()

	server.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("preflight status=%d body=%s", resp.Code, resp.Body.String())
	}
	if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "https://console.riido.io" {
		t.Fatalf("allow origin = %q", got)
	}
}
