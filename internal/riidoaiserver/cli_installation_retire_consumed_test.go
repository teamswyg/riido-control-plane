package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCLIInstallationRetirementPreservesConsumedReceipt(t *testing.T) {
	f := newCLIInstallationFixture(t)
	intent := f.intent(t)
	confirmed := f.post(t, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent)}, true)
	if confirmed.Code != http.StatusOK {
		t.Fatalf("confirm status=%d", confirmed.Code)
	}
	got := f.post(t, installationRetirePath, installationRetireBody(f, intent.IntentID), false)
	var result struct {
		Status  string                 `json:"status"`
		Receipt cliInstallationReceipt `json:"receipt"`
	}
	var link cliInstallationLink
	if got.Code != http.StatusOK || json.Unmarshal(got.Body.Bytes(), &result) != nil || json.Unmarshal(confirmed.Body.Bytes(), &link) != nil || result.Status != "consumed" || result.Receipt.IntentID != intent.IntentID || result.Receipt.Link != link {
		t.Fatalf("consumed retirement lost original receipt: status=%d", got.Code)
	}
}
