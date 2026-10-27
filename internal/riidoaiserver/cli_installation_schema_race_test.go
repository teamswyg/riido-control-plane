package riidoaiserver

import (
	"net/http"
	"testing"
	"time"
)

func TestCLIInstallationConfirmFencesConcurrentIntentSchemaChange(t *testing.T) {
	dynamo, newStore, _ := newTerminationDynamoFixture(t)
	dynamo.now = time.Now().UTC()
	f := newCLIInstallationFixture(t, StoreConfig{OperationStore: newStore()})
	intent := f.intent(t)
	dynamo.mu.Lock()
	dynamo.beforeCommit = func(items map[string]map[string]map[string]string) {
		key := terminationItemKey(cliInstallationIntentKey(intent.IntentID))
		items[key]["schema"] = map[string]string{"S": "unrecognized_schema/v2"}
	}
	dynamo.mu.Unlock()
	w := f.post(t, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent)}, true)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("changed intent schema was overwritten by stale consume: %d", w.Code)
	}
	dynamo.mu.Lock()
	defer dynamo.mu.Unlock()
	if len(dynamo.items[terminationItemKey(cliInstallationLinkKey(f.request.cliInstallationIdentity))]) > 0 {
		t.Fatal("schema race published a link")
	}
}
