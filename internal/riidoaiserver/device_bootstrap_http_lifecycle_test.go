package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestDeviceBootstrapHTTPPrepareActivateAndLostClientResponse(t *testing.T) {
	f := newBootstrapFixture(t)
	r := f.prepare(t)
	handler := f.handler(t, "cli-installation:bootstrap")
	proof := f.proof(t, r, "activate", strings.Repeat("b", 64))
	body := map[string]string{"schema_version": deviceBootstrapSchema, "intent_id": proof.IntentID, "boot_digest": proof.BootDigest, "signature": proof.Signature}
	path := "/v1/daemon/device-bootstraps/" + r.Request.TransitionID + "/activate"
	native := map[string]string{deviceIDHeader: f.credential.DeviceID, deviceSecretHeader: f.credential.DeviceSecret}
	if response := bootstrapHTTP(t, handler, path, body, native); response.Code != http.StatusOK {
		t.Fatalf("activate status %d", response.Code)
	}
	// The caller may have lost the response and still have the old primary
	// credential. Its encrypted journal contains the next secret already.
	status := "/v1/daemon/device-bootstraps/" + r.Request.TransitionID + "/status"
	if response := bootstrapHTTP(t, handler, status, map[string]string{"schema_version": deviceBootstrapSchema}, native); response.Code != http.StatusUnauthorized {
		t.Fatalf("retired native secret accepted: %d", response.Code)
	}
	native[deviceSecretHeader] = f.nextSecret
	response := bootstrapHTTP(t, handler, status, map[string]string{"schema_version": deviceBootstrapSchema}, native)
	if response.Code != http.StatusOK {
		t.Fatalf("journal recovery status %d", response.Code)
	}
	var wire deviceBootstrapResponse
	if err := json.Unmarshal(response.Body.Bytes(), &wire); err != nil {
		t.Fatal(err)
	}
	if wire.State != "complete" || wire.ExecutionCoverageGeneration != r.After.ExecutionCoverageGeneration || wire.CredentialGeneration != r.After.CredentialGeneration {
		t.Fatal("recovery replaced committed generations")
	}
}
