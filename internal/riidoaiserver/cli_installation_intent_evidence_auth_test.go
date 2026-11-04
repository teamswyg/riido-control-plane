package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCLIInstallationIntentEvidenceRejectsOtherBinding(t *testing.T) {
	f := newCLIInstallationFixture(t)
	original := installationIntentReplayBody(t, f)
	assertInstallationIntentReplayResponse(t, f.post(t, "/v1/internal/cli-installations/link-intents", original, false), original["request_id"])
	for _, key := range []string{"installation_id", "expected_device_id", "approved_principal_id", "resource", "client_id", "public_key", "key_fingerprint", "schema_version", "request_id"} {
		t.Run(key, func(t *testing.T) {
			body := installationEvidenceBody(original)
			body[key] = "other-valid-id"
			want := http.StatusConflict
			switch key {
			case "resource":
				body[key] = "https://other.fixture.invalid"
				want = http.StatusForbidden
			case "client_id", "public_key", "key_fingerprint", "schema_version", "request_id":
				want = http.StatusBadRequest
			}
			if got := f.post(t, installationEvidencePath, body, false); got.Code != want {
				t.Fatalf("changed %s status=%d want=%d", key, got.Code, want)
			}
		})
	}

	otherKey := newCLIInstallationFixture(t)
	otherBody := installationEvidenceBody(original)
	otherBody["public_key"], otherBody["key_fingerprint"] = otherKey.request.PublicKey, otherKey.request.KeyFingerprint
	if got := f.post(t, installationEvidencePath, otherBody, false); got.Code != http.StatusConflict {
		t.Fatal("valid replacement key read original evidence")
	}
	if got := f.post(t, installationEvidencePath, installationEvidenceBody(original), true); got.Code != http.StatusUnauthorized {
		t.Fatal("device credential accepted as service")
	}
	policyOnly, err := NewCLIInstallationServiceConfig(f.request.Resource, []StaticTokenCredential{{PrincipalID: cliInstallationServicePrincipal, Token: "service-fixture-token", Scopes: []string{"cli-installation:policy-read"}}})
	if err != nil {
		t.Fatal(err)
	}
	f.handler = NewServer(ServerConfig{Assignment: f.actor, AIAgentClient: f.devices, CLIInstallationService: policyOnly}).Handler()
	if got := f.post(t, installationEvidencePath, installationEvidenceBody(original), false); got.Code != http.StatusForbidden {
		t.Fatal("policy-only credential read intent evidence")
	}
}

func TestCLIInstallationIntentEvidenceRejectsMalformedAndBrowserRequests(t *testing.T) {
	f := newCLIInstallationFixture(t)
	for _, raw := range []string{`{`, `null`, `{}`, `{} {}`, `{"schema_version":"one","schema_version":"two"}`, `{"unknown":"private-sentinel"}`} {
		r := httptest.NewRequest(http.MethodPost, installationEvidencePath, strings.NewReader(raw))
		r.Header.Set("Authorization", "Bearer service-fixture-token")
		w := httptest.NewRecorder()
		f.handler.ServeHTTP(w, r)
		if w.Code != http.StatusBadRequest || strings.Contains(w.Body.String(), "private-sentinel") {
			t.Fatalf("malformed evidence status=%d", w.Code)
		}
	}
	for _, auth := range []string{"", "Bearer user-oauth-fixture"} {
		r := httptest.NewRequest(http.MethodPost, installationEvidencePath, strings.NewReader(`{}`))
		r.Header.Set("Authorization", auth)
		w := httptest.NewRecorder()
		f.handler.ServeHTTP(w, r)
		if w.Code != http.StatusUnauthorized {
			t.Fatal("missing or user authority accepted")
		}
	}
	r := httptest.NewRequest(http.MethodPost, installationEvidencePath, strings.NewReader(`{}`))
	r.Header.Set("Authorization", "Bearer service-fixture-token")
	r.Header.Set("Origin", "https://cli.fixture.invalid")
	w := httptest.NewRecorder()
	f.handler.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatal("browser authority accepted")
	}
}
