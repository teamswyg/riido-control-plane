package riidoaiserver

import (
	"net/http"
	"testing"
	"time"
)

func TestCLIInstallationImmutableLinkRejectsReplacementKey(t *testing.T) {
	for _, durable := range []bool{false, true} {
		t.Run(map[bool]string{false: "snapshot", true: "dynamo"}[durable], func(t *testing.T) {
			var configs []StoreConfig
			if durable {
				fixture, newStore, _ := newTerminationDynamoFixture(t)
				fixture.now = time.Now().UTC()
				configs = []StoreConfig{{OperationStore: newStore()}}
			}
			f := newCLIInstallationFixture(t, configs...)
			intent := f.intent(t)
			confirm := func(response cliInstallationIntentResponse) int {
				return f.post(t, "/v1/daemon/cli-installation-links/"+response.IntentID+"/confirm", map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, response)}, true).Code
			}
			if status := confirm(intent); status != http.StatusOK {
				t.Fatalf("first=%d", status)
			}
			other := newCLIInstallationFixture(t)
			f.key = other.key
			f.request.PublicKey = other.request.PublicKey
			f.request.KeyFingerprint = other.request.KeyFingerprint
			if status := confirm(f.intent(t)); status != http.StatusConflict {
				t.Fatalf("replacement key status=%d", status)
			}
		})
	}
}

func TestCLIInstallationExpiresUnconsumedNonceButReplaysOriginalACK(t *testing.T) {
	now := time.Now().UTC()
	f := newCLIInstallationFixture(t, StoreConfig{Now: func() time.Time { return now }})
	first, expired := f.intent(t), f.intent(t)
	body := map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, first)}
	path := "/v1/daemon/cli-installation-links/" + first.IntentID + "/confirm"
	before := f.post(t, path, body, true)
	if before.Code != http.StatusOK {
		t.Fatalf("initial=%d", before.Code)
	}
	now = first.ExpiresAt.Add(time.Minute)
	after := f.post(t, path, body, true)
	if after.Code != http.StatusOK || before.Body.String() != after.Body.String() {
		t.Fatal("consumed nonce lost immutable ACK after expiry")
	}
	w := f.post(t, "/v1/daemon/cli-installation-links/"+expired.IntentID+"/confirm", map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, expired)}, true)
	if w.Code != http.StatusGone {
		t.Fatalf("expired unconsumed nonce=%d", w.Code)
	}
}
