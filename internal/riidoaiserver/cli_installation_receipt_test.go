package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func installationReceiptBody(f *cliInstallationFixture, id string) map[string]string {
	r := f.request
	return map[string]string{
		"schema_version": "riido-cli-installation-link-receipt.v1", "intent_id": id,
		"installation_id": r.InstallationID, "resource": r.Resource,
		"client_id": r.ClientID, "key_fingerprint": r.KeyFingerprint,
		"expected_device_id": r.ExpectedDeviceID, "approved_principal_id": r.ApprovedPrincipalID,
	}
}

func TestCLIInstallationReceiptRequiresOriginalConsumedIntent(t *testing.T) {
	f := newCLIInstallationFixture(t)
	intent := f.intent(t)
	request := installationReceiptBody(f, intent.IntentID)
	path := "/v1/internal/cli-installations/link-receipts"
	if got := f.post(t, path, request, false); got.Code != http.StatusConflict {
		t.Fatalf("unconsumed receipt status=%d", got.Code)
	}
	confirmed := f.post(t, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", map[string]string{
		"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent),
	}, true)
	if confirmed.Code != http.StatusOK {
		t.Fatalf("confirm status=%d", confirmed.Code)
	}
	got := f.post(t, path, request, false)
	var receipt struct {
		SchemaVersion string              `json:"schema_version"`
		IntentID      string              `json:"intent_id"`
		Link          cliInstallationLink `json:"link"`
	}
	var original cliInstallationLink
	if got.Code != http.StatusOK || json.Unmarshal(got.Body.Bytes(), &receipt) != nil || json.Unmarshal(confirmed.Body.Bytes(), &original) != nil {
		t.Fatalf("original receipt status=%d", got.Code)
	}
	if receipt.SchemaVersion != request["schema_version"] || receipt.IntentID != intent.IntentID || receipt.Link != original {
		t.Fatal("receipt lost original intent or immutable ACK")
	}
	other := f.intent(t)
	if got := f.post(t, path, installationReceiptBody(f, other.IntentID), false); got.Code != http.StatusConflict {
		t.Fatalf("unconsumed intent borrowed existing link: status=%d", got.Code)
	}
	if got := f.post(t, path, request, true); got.Code != http.StatusUnauthorized {
		t.Fatalf("device credential replaced service authentication: status=%d", got.Code)
	}
}

func TestCLIInstallationReceiptRejectsMismatchedBinding(t *testing.T) {
	f := newCLIInstallationFixture(t)
	intent := f.intent(t)
	for _, key := range []string{"intent_id", "installation_id", "client_id", "expected_device_id", "approved_principal_id"} {
		t.Run(key, func(t *testing.T) {
			request := installationReceiptBody(f, intent.IntentID)
			request[key] = "other-valid-id"
			want := http.StatusForbidden
			if key == "client_id" {
				want = http.StatusBadRequest
			}
			if got := f.post(t, "/v1/internal/cli-installations/link-receipts", request, false); got.Code != want {
				t.Fatalf("mismatched identity status=%d", got.Code)
			}
		})
	}
}
