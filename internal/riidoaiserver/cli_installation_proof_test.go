package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCLIInstallationLinkRequiresApprovedDeviceOwnerAndKey(t *testing.T) {
	for _, which := range []string{"approved-principal", "device", "signature", "key"} {
		t.Run(which, func(t *testing.T) {
			f := newCLIInstallationFixture(t)
			if which == "approved-principal" {
				f.request.ApprovedPrincipalID = "different-user"
			}
			if which == "device" {
				f.request.ExpectedDeviceID = "different-device"
			}
			intent := f.intent(t)
			signature := f.signature(t, intent)
			if which == "signature" {
				signature = "Zm9yZ2Vk"
			}
			if which == "key" {
				other := newCLIInstallationFixture(t)
				other.request = f.request
				signature = other.signature(t, intent)
			}
			w := f.post(t, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": signature}, true)
			if w.Code != http.StatusForbidden {
				t.Fatalf("bad proof status=%d", w.Code)
			}
		})
	}
}

func TestCLIInstallationOwnerChangeNeverRestoresPermission(t *testing.T) {
	f := newCLIInstallationFixture(t)
	intent := f.intent(t)
	body := map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent)}
	path := "/v1/daemon/cli-installation-links/" + intent.IntentID + "/confirm"
	if w := f.post(t, path, body, true); w.Code != http.StatusOK {
		t.Fatalf("confirm=%d", w.Code)
	}
	f.devices.mu.Lock()
	device := f.devices.deviceCredentials["fixture-device"]
	device.ownerPrincipalID = "changed-owner"
	f.devices.deviceCredentials["fixture-device"] = device
	f.devices.mu.Unlock()
	if w := f.post(t, path, body, true); w.Code != http.StatusForbidden {
		t.Fatalf("old owner receipt=%d", w.Code)
	}
	w := f.post(t, "/v1/internal/cli-installations/policy", map[string]string{"schema_version": cliInstallationPolicySchema, "installation_id": f.request.InstallationID, "resource": f.request.Resource, "client_id": f.request.ClientID, "key_fingerprint": f.request.KeyFingerprint}, false)
	var policy cliInstallationPolicy
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &policy) != nil || policy.Decision != "unknown" {
		t.Fatalf("changed owner policy=%d %+v", w.Code, policy)
	}
}
