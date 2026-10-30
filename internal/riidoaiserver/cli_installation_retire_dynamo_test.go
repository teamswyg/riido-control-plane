package riidoaiserver

import (
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestCLIInstallationRetirementDynamoFailureLostACKRestartAndReplay(t *testing.T) {
	dynamo, newStore, _ := newTerminationDynamoFixture(t)
	var clock atomic.Int64
	clock.Store(time.Now().UTC().UnixNano())
	now := func() time.Time { return time.Unix(0, clock.Load()).UTC() }
	operations := newStore()
	operations.now = now
	f := newCLIInstallationFixture(t, StoreConfig{OperationStore: operations})
	replay := installationIntentReplayBody(t, f)
	created := f.post(t, "/v1/internal/cli-installations/link-intents", replay, false)
	intent := assertInstallationIntentReplayResponse(t, created, replay["request_id"])
	clock.Store(intent.ExpiresAt.UnixNano())
	body := installationRetireBody(f, intent.IntentID)
	dynamo.mu.Lock()
	dynamo.deny = true
	dynamo.mu.Unlock()
	if got := f.post(t, installationRetirePath, body, false); got.Code != http.StatusServiceUnavailable {
		t.Fatalf("failed retirement acknowledged: status=%d", got.Code)
	}
	dynamo.mu.Lock()
	dynamo.deny, dynamo.loseACK = false, true
	dynamo.mu.Unlock()
	first := f.post(t, installationRetirePath, body, false)
	if first.Code != http.StatusOK || !strings.Contains(first.Body.String(), `"status":"retired"`) {
		t.Fatalf("lost ACK not recovered: status=%d", first.Code)
	}
	reopenedOperations := newStore()
	reopenedOperations.now = now
	reopened := NewStoreWithConfig(StoreConfig{OperationStore: reopenedOperations})
	t.Cleanup(reopened.Close)
	f.handler = NewServer(ServerConfig{Assignment: reopened, AIAgentClient: f.devices, CLIInstallationService: f.service}).Handler()
	if got := f.post(t, installationRetirePath, body, false); got.Code != http.StatusOK || got.Body.String() != first.Body.String() {
		t.Fatalf("restart lost retirement: status=%d", got.Code)
	}
	if got := f.post(t, "/v1/internal/cli-installations/link-intents", replay, false); got.Code != http.StatusOK || got.Body.String() != created.Body.String() {
		t.Fatalf("retirement rewrote original replay: status=%d", got.Code)
	}
	clock.Store(intent.ExpiresAt.Add(-time.Second).UnixNano())
	if got := f.post(t, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent)}, true); got.Code != http.StatusGone {
		t.Fatalf("replayed intent reopened: status=%d", got.Code)
	}
	if got := f.post(t, "/v1/internal/cli-installations/link-receipts", installationReceiptBody(f, intent.IntentID), false); got.Code != http.StatusGone {
		t.Fatalf("retirement fabricated consumed receipt: status=%d", got.Code)
	}
}
