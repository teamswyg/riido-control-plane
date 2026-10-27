package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestCLIInstallationDynamoAtomicLinkLostACKAndRestart(t *testing.T) {
	dynamo, newStore, _ := newTerminationDynamoFixture(t)
	dynamo.now = time.Now().UTC()
	f := newCLIInstallationFixture(t, StoreConfig{OperationStore: newStore()})
	intent := f.intent(t)
	body := map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent)}
	path := "/v1/daemon/cli-installation-links/" + intent.IntentID + "/confirm"
	dynamo.mu.Lock()
	dynamo.deny = true
	dynamo.mu.Unlock()
	if w := f.post(t, path, body, true); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("failed transaction status=%d", w.Code)
	}
	dynamo.mu.Lock()
	if len(dynamo.items[terminationItemKey(cliInstallationLinkKey(f.request.cliInstallationIdentity))]) != 0 || len(dynamo.items[terminationItemKey(devicePolicyKey("fixture-device"))]) != 0 {
		t.Fatal("failed confirm published partial link/revision")
	}
	dynamo.deny, dynamo.loseACK = false, true
	dynamo.mu.Unlock()
	first := f.post(t, path, body, true)
	if first.Code != http.StatusOK {
		t.Fatalf("lost ACK retry=%d", first.Code)
	}
	reopened := NewStoreWithConfig(StoreConfig{OperationStore: newStore()})
	t.Cleanup(reopened.Close)
	f.handler = NewServer(ServerConfig{Assignment: reopened, AIAgentClient: f.devices, CLIInstallationService: f.service}).Handler()
	again := f.post(t, path, body, true)
	if again.Code != http.StatusOK || again.Body.String() != first.Body.String() {
		t.Fatal("reopened authority lost original link ACK")
	}
	policy, err := reopened.readCLIInstallationPolicy(t.Context(), f.request.cliInstallationIdentity, f.devices)
	if err != nil || policy.Decision != "unknown" || policy.ExecutionRevision != "1" {
		t.Fatalf("empty history/retry became allowed or advanced twice: %+v %v", policy, err)
	}
	dynamo.mu.Lock()
	defer dynamo.mu.Unlock()
	var stored cliInstallationIntent
	item := dynamo.items[terminationItemKey(cliInstallationIntentKey(intent.IntentID))]
	if json.Unmarshal([]byte(item["payload"]["S"]), &stored) != nil || !stored.Consumed {
		t.Fatal("nonce was not consumed with link")
	}
}
