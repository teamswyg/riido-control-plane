package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCLIInstallationConcurrentConfirmationConsumesOnce(t *testing.T) {
	dynamo, newStore, _ := newTerminationDynamoFixture(t)
	dynamo.now = time.Now().UTC()
	f := newCLIInstallationFixture(t, StoreConfig{OperationStore: newStore()})
	intent := f.intent(t)
	body := map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent)}
	responses := make(chan *httptest.ResponseRecorder, 2)
	for range 2 {
		go func() {
			responses <- f.post(t, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", body, true)
		}()
	}
	one, two := <-responses, <-responses
	if one.Code != http.StatusOK || two.Code != http.StatusOK || one.Body.String() != two.Body.String() {
		t.Fatalf("concurrent confirmation lost immutable ACK: %d/%d", one.Code, two.Code)
	}
	policy, err := f.actor.readCLIInstallationPolicy(t.Context(), f.request.cliInstallationIdentity, f.devices)
	if err != nil || policy.ExecutionRevision != "1" || policy.Decision != "unknown" {
		t.Fatalf("nonce consumed more than once: %+v %v", policy, err)
	}
}
