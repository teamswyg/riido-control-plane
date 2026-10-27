package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCLIInstallationServiceCannotUseUserWildcardOrDevice(t *testing.T) {
	for _, scope := range []string{"riido:*", "cli-installation:*", "metrics:read"} {
		_, err := NewCLIInstallationServiceConfig("https://cli.fixture.invalid", []StaticTokenCredential{{PrincipalID: cliInstallationServicePrincipal, Token: "fixture", Scopes: []string{scope}}})
		if err == nil {
			t.Fatalf("service wildcard accepted: %s", scope)
		}
	}
	f := newCLIInstallationFixture(t)
	for _, mode := range []string{"user", "device", "mixed", "duplicate", "empty"} {
		t.Run(mode, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/v1/internal/cli-installations/policy", strings.NewReader(`{}`))
			switch mode {
			case "user":
				r.Header.Set("Authorization", "Bearer user-fixture")
			case "device":
				r.Header.Set(deviceIDHeader, "fixture-device")
				r.Header.Set(deviceSecretHeader, "fixture-device-secret")
			case "mixed":
				r.Header.Set("Authorization", "Bearer service-fixture-token")
				r.Header.Set(deviceIDHeader, "fixture-device")
			case "duplicate":
				r.Header.Add("Authorization", "Bearer service-fixture-token")
				r.Header.Add("Authorization", "Bearer service-fixture-token")
			}
			w := httptest.NewRecorder()
			f.handler.ServeHTTP(w, r)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status=%d", w.Code)
			}
		})
	}
}

func TestCLIInstallationStrictJSONAndResource(t *testing.T) {
	f := newCLIInstallationFixture(t)
	for _, body := range []string{`{`, `null`, `{} {}`, `{"schema_version":"one","schema_version":"two"}`, `{"owner":"secret-sentinel"}`} {
		r := httptest.NewRequest(http.MethodPost, "/v1/internal/cli-installations/link-intents", strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer service-fixture-token")
		w := httptest.NewRecorder()
		f.handler.ServeHTTP(w, r)
		if w.Code != http.StatusBadRequest || strings.Contains(w.Body.String(), "secret-sentinel") {
			t.Fatalf("malformed status=%d", w.Code)
		}
	}
	r := httptest.NewRequest(http.MethodPost, "/v1/internal/cli-installations/link-intents", strings.NewReader(`{"schema_version":"`+strings.Repeat("x", 5000)+`"}`))
	r.Header.Set("Authorization", "Bearer service-fixture-token")
	w := httptest.NewRecorder()
	f.handler.ServeHTTP(w, r)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversize status=%d", w.Code)
	}
	f.request.Resource = "https://other.fixture.invalid"
	if w = f.post(t, "/v1/internal/cli-installations/link-intents", f.request, false); w.Code != http.StatusForbidden {
		t.Fatalf("cross-resource=%d", w.Code)
	}
}
