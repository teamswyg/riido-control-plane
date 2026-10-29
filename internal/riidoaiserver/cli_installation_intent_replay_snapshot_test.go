package riidoaiserver

import (
	"net/http"
	"path/filepath"
	"testing"
)

func TestCLIInstallationIntentV2SnapshotFailureAndV1Restart(t *testing.T) {
	snapshot, err := NewFileStoreSnapshot(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	failing := &failingClaimSnapshot{SnapshotStore: snapshot}
	f := newCLIInstallationFixture(t, StoreConfig{SnapshotStore: failing})
	legacy := f.intent(t)
	f.actor.Close()
	legacyStore, err := OpenStoreWithConfig(t.Context(), StoreConfig{SnapshotStore: failing})
	if err != nil {
		t.Fatalf("v1-only snapshot no longer restores: %v", err)
	}
	t.Cleanup(legacyStore.Close)
	f.handler = NewServer(ServerConfig{Assignment: legacyStore, AIAgentClient: f.devices, CLIInstallationService: f.service}).Handler()
	confirm := map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, legacy)}
	if got := f.post(t, "/v1/daemon/cli-installation-links/"+legacy.IntentID+"/confirm", confirm, true); got.Code != http.StatusOK {
		t.Fatalf("restored legacy intent failed confirmation: %d", got.Code)
	}
	body := installationIntentReplayBody(t, f)
	failing.fail = true
	if got := f.post(t, "/v1/internal/cli-installations/link-intents", body, false); got.Code != http.StatusServiceUnavailable {
		t.Fatalf("failed snapshot status=%d", got.Code)
	}
	failing.fail = false
	first := f.post(t, "/v1/internal/cli-installations/link-intents", body, false)
	assertInstallationIntentReplayResponse(t, first, body["request_id"])
	legacyStore.Close()
	reopened, err := OpenStoreWithConfig(t.Context(), StoreConfig{SnapshotStore: snapshot})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(reopened.Close)
	f.handler = NewServer(ServerConfig{Assignment: reopened, AIAgentClient: f.devices, CLIInstallationService: f.service}).Handler()
	again := f.post(t, "/v1/internal/cli-installations/link-intents", body, false)
	if again.Code != http.StatusOK || again.Body.String() != first.Body.String() {
		t.Fatal("snapshot restart changed original response or published a failed save")
	}
}
