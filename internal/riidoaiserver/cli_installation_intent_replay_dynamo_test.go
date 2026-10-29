package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"
)

func TestCLIInstallationIntentV2DynamoLostACKAndRestart(t *testing.T) {
	dynamo, newStore, _ := newTerminationDynamoFixture(t)
	f := newCLIInstallationFixture(t, StoreConfig{OperationStore: newStore()})
	body := installationIntentReplayBody(t, f)
	dynamo.mu.Lock()
	dynamo.deny = true
	dynamo.mu.Unlock()
	if got := f.post(t, "/v1/internal/cli-installations/link-intents", body, false); got.Code != http.StatusServiceUnavailable {
		t.Fatalf("transaction failure status=%d", got.Code)
	}
	dynamo.mu.Lock()
	partial := false
	for key := range dynamo.items {
		if strings.Contains(key, "CLI_INSTALLATION") {
			partial = true
		}
	}
	dynamo.deny, dynamo.loseACK = false, true
	dynamo.mu.Unlock()
	if partial {
		t.Fatal("failed generation published partial authority")
	}
	first := f.post(t, "/v1/internal/cli-installations/link-intents", body, false)
	assertInstallationIntentReplayResponse(t, first, body["request_id"])
	reopened := NewStoreWithConfig(StoreConfig{OperationStore: newStore()})
	t.Cleanup(reopened.Close)
	f.handler = NewServer(ServerConfig{Assignment: reopened, AIAgentClient: f.devices, CLIInstallationService: f.service}).Handler()
	again := f.post(t, "/v1/internal/cli-installations/link-intents", body, false)
	if again.Code != http.StatusOK || again.Body.String() != first.Body.String() {
		t.Fatal("Dynamo restart did not recover original generation response")
	}
}
