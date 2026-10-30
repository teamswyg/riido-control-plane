package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCLIInstallationRetirementRejectsOtherIdentityAndAuthority(t *testing.T) {
	f := newCLIInstallationFixture(t)
	intent := f.intent(t)
	for _, key := range []string{"intent_id", "installation_id", "expected_device_id", "approved_principal_id", "resource", "client_id", "key_fingerprint"} {
		t.Run(key, func(t *testing.T) {
			body := installationRetireBody(f, intent.IntentID)
			body[key] = "other-valid-id"
			if key == "resource" {
				body[key] = "https://other.fixture.invalid"
			}
			if key == "key_fingerprint" {
				body[key] = strings.Repeat("0", 64)
			}
			want := http.StatusForbidden
			if key == "client_id" {
				want = http.StatusBadRequest
			}
			if got := f.post(t, installationRetirePath, body, false); got.Code != want {
				t.Fatalf("cross-binding retirement %s status=%d", key, got.Code)
			}
		})
	}
	if got := f.post(t, installationRetirePath, installationRetireBody(f, intent.IntentID), true); got.Code != http.StatusUnauthorized {
		t.Fatalf("device credential retired intent: %d", got.Code)
	}
	policyOnly, err := NewCLIInstallationServiceConfig(f.request.Resource, []StaticTokenCredential{{PrincipalID: cliInstallationServicePrincipal, Token: "service-fixture-token", Scopes: []string{"cli-installation:policy-read"}}})
	if err != nil {
		t.Fatal(err)
	}
	f.handler = NewServer(ServerConfig{Assignment: f.actor, AIAgentClient: f.devices, CLIInstallationService: policyOnly}).Handler()
	if got := f.post(t, installationRetirePath, installationRetireBody(f, intent.IntentID), false); got.Code != http.StatusForbidden {
		t.Fatalf("read-only service retired intent: %d", got.Code)
	}
}

func TestCLIInstallationRetirementRejectsBrowserAndMalformedJSON(t *testing.T) {
	f := newCLIInstallationFixture(t)
	for _, raw := range []string{`{`, `null`, `{}`, `{} {}`, `{"schema_version":"one","schema_version":"two"}`, `{"unknown":"private-sentinel"}`} {
		r := httptest.NewRequest(http.MethodPost, installationRetirePath, strings.NewReader(raw))
		r.Header.Set("Authorization", "Bearer service-fixture-token")
		w := httptest.NewRecorder()
		f.handler.ServeHTTP(w, r)
		if w.Code != http.StatusBadRequest || strings.Contains(w.Body.String(), "private-sentinel") {
			t.Fatalf("malformed retirement status=%d", w.Code)
		}
	}
	r := httptest.NewRequest(http.MethodPost, installationRetirePath, strings.NewReader(`{}`))
	r.Header.Set("Authorization", "Bearer service-fixture-token")
	r.Header.Set("Origin", "https://cli.fixture.invalid")
	w := httptest.NewRecorder()
	f.handler.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatalf("browser request retired intent: %d", w.Code)
	}
}
