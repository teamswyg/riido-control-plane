package riidoaiserver

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCLIInstallationRetirementConsumedResponseFitsExistingBudget(t *testing.T) {
	for _, durable := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "dynamo"}[durable], func(t *testing.T) {
			var config StoreConfig
			if durable {
				_, newStore, _ := newTerminationDynamoFixture(t)
				operations := newStore()
				operations.now = func() time.Time { return time.Now().UTC() }
				config.OperationStore = operations
			}
			f := newCLIInstallationFixture(t, config)
			largeID := strings.Repeat("\"", 512)
			f.request.InstallationID, f.request.ExpectedDeviceID = largeID, largeID
			f.devices.deviceCredentials[largeID] = deviceCredentialRecord{deviceID: largeID, ownerPrincipalID: "owner", secretHash: sha256.Sum256([]byte("fixture-device-secret"))}
			intent := f.intent(t) // Accepted public input and intent storage bounds.
			body, err := json.Marshal(map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent)})
			if err != nil {
				t.Fatal(err)
			}
			r := httptest.NewRequest(http.MethodPost, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", bytes.NewReader(body))
			r.Header.Set(deviceIDHeader, largeID)
			r.Header.Set(deviceSecretHeader, "fixture-device-secret")
			confirmed := httptest.NewRecorder()
			f.handler.ServeHTTP(confirmed, r)
			if confirmed.Code != http.StatusOK {
				t.Fatalf("large valid confirmation status=%d", confirmed.Code)
			}
			receipt := f.post(t, "/v1/internal/cli-installations/link-receipts", installationReceiptBody(f, intent.IntentID), false)
			if receipt.Code != http.StatusOK {
				t.Fatalf("ordinary receipt status=%d", receipt.Code)
			}
			got := f.post(t, installationRetirePath, installationRetireBody(f, intent.IntentID), false)
			var result cliInstallationRetirement
			var original cliInstallationReceipt
			if got.Code != http.StatusOK || got.Body.Len() > (4<<10)+1 || json.Unmarshal(got.Body.Bytes(), &result) != nil || json.Unmarshal(receipt.Body.Bytes(), &original) != nil || result.Status != "consumed" || result.Receipt == nil || *result.Receipt != original {
				t.Fatalf("valid consumed retirement lost original receipt: status=%d bytes=%d", got.Code, got.Body.Len())
			}
		})
	}
}
