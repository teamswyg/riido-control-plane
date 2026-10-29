package riidoaiserver

import (
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestCLIInstallationIntentV2ReplayDoesNotRenewExpiry(t *testing.T) {
	for _, durable := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "dynamo"}[durable], func(t *testing.T) {
			var later atomic.Pointer[time.Time]
			clock := func() time.Time {
				if now := later.Load(); now != nil {
					return *now
				}
				return time.Now().UTC()
			}
			config := StoreConfig{Now: clock}
			if durable {
				_, newStore, _ := newTerminationDynamoFixture(t)
				operations := newStore()
				operations.now = clock
				config.OperationStore = operations
			}
			f := newCLIInstallationFixture(t, config)
			body := installationIntentReplayBody(t, f)
			first := f.post(t, "/v1/internal/cli-installations/link-intents", body, false)
			intent := assertInstallationIntentReplayResponse(t, first, body["request_id"])
			after := intent.ExpiresAt.Add(time.Second)
			later.Store(&after)
			again := f.post(t, "/v1/internal/cli-installations/link-intents", body, false)
			if again.Code != http.StatusOK || again.Body.String() != first.Body.String() {
				t.Fatal("expired retry did not preserve original response")
			}
			confirm := map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent)}
			if got := f.post(t, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", confirm, true); got.Code != http.StatusGone {
				t.Fatalf("replay renewed expired authority: %d", got.Code)
			}
		})
	}
}
