package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCLIInstallationReceiptRejectsWrongScopeAndMalformedJSON(t *testing.T) {
	f := newCLIInstallationFixture(t)
	intent := f.intent(t)
	policyOnly, err := NewCLIInstallationServiceConfig(f.request.Resource, []StaticTokenCredential{{
		PrincipalID: cliInstallationServicePrincipal, Token: "service-fixture-token", Scopes: []string{"cli-installation:policy-read"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	f.handler = NewServer(ServerConfig{Assignment: f.actor, AIAgentClient: f.devices, CLIInstallationService: policyOnly}).Handler()
	path := "/v1/internal/cli-installations/link-receipts"
	if got := f.post(t, path, installationReceiptBody(f, intent.IntentID), false); got.Code != http.StatusForbidden {
		t.Fatalf("policy scope recovered link receipt: status=%d", got.Code)
	}
	f.handler = NewServer(ServerConfig{Assignment: f.actor, AIAgentClient: f.devices, CLIInstallationService: f.service}).Handler()
	for _, raw := range []string{`{`, `null`, `{}`, `{} {}`, `{"schema_version":"one","schema_version":"two"}`, `{"unknown":"private-sentinel"}`} {
		r := httptest.NewRequest(http.MethodPost, path, strings.NewReader(raw))
		r.Header.Set("Authorization", "Bearer service-fixture-token")
		w := httptest.NewRecorder()
		f.handler.ServeHTTP(w, r)
		if w.Code != http.StatusBadRequest || strings.Contains(w.Body.String(), "private-sentinel") {
			t.Fatalf("malformed receipt status=%d", w.Code)
		}
	}
	for _, key := range []string{"resource", "key_fingerprint"} {
		body := installationReceiptBody(f, intent.IntentID)
		body[key] = "https://other.fixture.invalid"
		if key == "key_fingerprint" {
			body[key] = strings.Repeat("0", 64)
		}
		if got := f.post(t, path, body, false); got.Code != http.StatusForbidden {
			t.Fatalf("cross-binding receipt %s status=%d", key, got.Code)
		}
	}
}
