package riidoaiserver

import (
	"net/http"
	"testing"
	"time"
)

func TestCLIInstallationDynamoPolicyRejectsMissingOrCorruptRevision(t *testing.T) {
	for _, revision := range []string{"", "0", "01", "18446744073709551616"} {
		t.Run(revision, func(t *testing.T) {
			dynamo, newStore, _ := newTerminationDynamoFixture(t)
			dynamo.now = time.Now().UTC()
			f := newCLIInstallationFixture(t, StoreConfig{OperationStore: newStore()})
			intent := f.intent(t)
			if w := f.post(t, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent)}, true); w.Code != http.StatusOK {
				t.Fatalf("confirm=%d", w.Code)
			}
			dynamo.mu.Lock()
			key := terminationItemKey(devicePolicyKey("fixture-device"))
			if revision == "" {
				delete(dynamo.items, key)
			} else {
				dynamo.items[key]["revision"] = map[string]string{"N": revision}
			}
			dynamo.mu.Unlock()
			if _, err := f.actor.readCLIInstallationPolicy(t.Context(), f.request.cliInstallationIdentity, f.devices); err == nil {
				t.Fatal("invalid authority inferred permission")
			}
		})
	}
}
