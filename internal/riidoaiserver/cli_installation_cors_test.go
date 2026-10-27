package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCLIInstallationPrivateRoutesRejectBrowserOrigin(t *testing.T) {
	f := newCLIInstallationFixture(t)
	handler := NewServer(ServerConfig{Assignment: f.actor, AIAgentClient: f.devices, CLIInstallationService: f.service, WebAllowedOrigins: []string{"https://web.fixture.invalid"}}).Handler()
	for _, path := range []string{"/v1/internal/cli-installations/link-intents", "/v1/internal/cli-installations/policy", "/v1/daemon/cli-installation-links/intent/confirm"} {
		for _, method := range []string{http.MethodOptions, http.MethodPost} {
			r := httptest.NewRequest(method, path, strings.NewReader(`{}`))
			r.Header.Set("Origin", "https://web.fixture.invalid")
			r.Header.Set("Authorization", "Bearer service-fixture-token")
			if method == http.MethodOptions {
				r.Header.Set("Access-Control-Request-Method", http.MethodPost)
				r.Header.Set("Access-Control-Request-Headers", "Authorization, Content-Type")
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code != http.StatusForbidden || w.Header().Get("Access-Control-Allow-Origin") != "" {
				t.Fatalf("private route became browser surface: %s %s=%d", method, path, w.Code)
			}
		}
	}
}
