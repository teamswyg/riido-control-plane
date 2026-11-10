package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestDeviceBootstrapHTTPScopesAndCredentialIsolation(t *testing.T) {
	f := newBootstrapFixture(t)
	request := f.request
	request.SchemaVersion = deviceBootstrapSchema
	path := "/v1/internal/cli-installations/device-bootstraps"
	service := map[string]string{"Authorization": "Bearer bootstrap-fixture-service"}
	for _, scope := range []string{"cli-installation:link", "cli-installation:transfer", "cli-installation:policy-read"} {
		if got := bootstrapHTTP(t, f.handler(t, scope), path, request, service); got.Code != http.StatusForbidden {
			t.Fatalf("scope %s status %d", scope, got.Code)
		}
	}
	handler := f.handler(t, "cli-installation:bootstrap")
	response := bootstrapHTTP(t, handler, path, request, service)
	if response.Code != http.StatusOK {
		t.Fatalf("create status %d", response.Code)
	}
	var wire deviceBootstrapResponse
	if err := json.Unmarshal(response.Body.Bytes(), &wire); err != nil {
		t.Fatal(err)
	}
	if wire.TransitionID != request.TransitionID || wire.State != "issued" {
		t.Fatal("wrong created transition")
	}
	for _, forbidden := range []string{"secret_hash", "next_secret", "public_key", "before", "after", "credential_issued_at"} {
		if strings.Contains(response.Body.String(), forbidden) {
			t.Fatal("HTTP exposed persistence authority", forbidden)
		}
	}
	statusPath := "/v1/daemon/device-bootstraps/" + request.TransitionID + "/status"
	body := map[string]string{"schema_version": deviceBootstrapSchema}
	if got := bootstrapHTTP(t, handler, statusPath, body, service); got.Code != http.StatusUnauthorized {
		t.Fatalf("service fallback accepted: %d", got.Code)
	}
	native := map[string]string{deviceIDHeader: f.credential.DeviceID, deviceSecretHeader: f.credential.DeviceSecret}
	if got := bootstrapHTTP(t, handler, statusPath, body, native); got.Code != http.StatusOK {
		t.Fatalf("native status %d", got.Code)
	}
	native["Origin"] = "https://attacker.invalid"
	if got := bootstrapHTTP(t, handler, statusPath, body, native); got.Code != http.StatusForbidden {
		t.Fatalf("browser origin accepted: %d", got.Code)
	}
}
