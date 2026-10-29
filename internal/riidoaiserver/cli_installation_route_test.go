package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCLIInstallationRoutesFailClosedWithoutAuthority(t *testing.T) {
	handler := NewServer(ServerConfig{}).Handler()
	for _, tc := range []struct {
		path   string
		status int
	}{
		{"/v1/internal/cli-installations/link-intents", http.StatusServiceUnavailable},
		{"/v1/internal/cli-installations/policy", http.StatusServiceUnavailable},
		{"/v1/internal/cli-installations/link-receipts", http.StatusServiceUnavailable},
		{"/v1/daemon/cli-installation-links/fixture/confirm", http.StatusUnauthorized},
	} {
		t.Run(tc.path, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(`{}`))
			r.Header.Set("Authorization", "Bearer user-token-sentinel")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Fatalf("private route status=%d want=%d", w.Code, tc.status)
			}
			if strings.Contains(w.Body.String(), "user-token-sentinel") {
				t.Fatal("route reflected credential")
			}
		})
	}
}
