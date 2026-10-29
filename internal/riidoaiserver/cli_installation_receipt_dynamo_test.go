package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestCLIInstallationReceiptDynamoRestartAndMissingLink(t *testing.T) {
	dynamo, newStore, _ := newTerminationDynamoFixture(t)
	dynamo.now = time.Now().UTC()
	f := newCLIInstallationFixture(t, StoreConfig{OperationStore: newStore()})
	intent := f.intent(t)
	confirmed := f.post(t, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", map[string]string{
		"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent),
	}, true)
	if confirmed.Code != http.StatusOK {
		t.Fatalf("confirm status=%d", confirmed.Code)
	}
	reopened := NewStoreWithConfig(StoreConfig{OperationStore: newStore()})
	t.Cleanup(reopened.Close)
	f.handler = NewServer(ServerConfig{Assignment: reopened, AIAgentClient: f.devices, CLIInstallationService: f.service}).Handler()
	body := installationReceiptBody(f, intent.IntentID)
	path := "/v1/internal/cli-installations/link-receipts"
	got := f.post(t, path, body, false)
	var receipt cliInstallationReceipt
	var original cliInstallationLink
	if got.Code != http.StatusOK || json.Unmarshal(got.Body.Bytes(), &receipt) != nil || json.Unmarshal(confirmed.Body.Bytes(), &original) != nil || receipt.Link != original {
		t.Fatalf("durable receipt lost original ACK: status=%d", got.Code)
	}
	// Structurally valid storage must not turn an impossible confirmation time
	// into an authenticated receipt that the consumer cannot accept.
	for name, linkedAt := range map[string]time.Time{"after": intent.ExpiresAt.Add(time.Nanosecond), "equal": intent.ExpiresAt} {
		t.Run(name, func(t *testing.T) {
			damaged := original
			damaged.LinkedAt = linkedAt
			item, err := cliInstallationItem(cliInstallationLinkKey(f.request.cliInstallationIdentity), cliInstallationLinkItemSchema, damaged)
			if err != nil {
				t.Fatal(err)
			}
			dynamo.mu.Lock()
			dynamo.items[terminationItemKey(cliInstallationLinkKey(f.request.cliInstallationIdentity))] = item
			dynamo.mu.Unlock()
			if got := f.post(t, path, body, false); got.Code != http.StatusServiceUnavailable {
				t.Fatalf("link confirmed at or after intent expiry succeeded: status=%d", got.Code)
			}
		})
	}
	dynamo.mu.Lock()
	delete(dynamo.items, terminationItemKey(cliInstallationLinkKey(f.request.cliInstallationIdentity)))
	dynamo.mu.Unlock()
	if got := f.post(t, path, body, false); got.Code != http.StatusServiceUnavailable {
		t.Fatalf("consumed without link succeeded: status=%d", got.Code)
	}
}
