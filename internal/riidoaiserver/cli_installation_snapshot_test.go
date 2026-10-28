package riidoaiserver

import (
	"net/http"
	"path/filepath"
	"testing"
)

func TestCLIInstallationSnapshotFailureAndRestart(t *testing.T) {
	snapshot, err := NewFileStoreSnapshot(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	failing := &failingClaimSnapshot{SnapshotStore: snapshot}
	f := newCLIInstallationFixture(t, StoreConfig{SnapshotStore: failing})
	intent := f.intent(t)
	path := "/v1/daemon/cli-installation-links/" + intent.IntentID + "/confirm"
	body := map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent)}
	failing.fail = true
	if w := f.post(t, path, body, true); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("failed persistence=%d", w.Code)
	}
	failing.fail = false
	result, err := f.actor.executeCLIInstallation(t.Context(), cliInstallationCommand{read: &cliInstallationRead{identity: f.request.cliInstallationIdentity}})
	if err != nil || result.read.found {
		t.Fatal("failed snapshot published link")
	}
	first := f.post(t, path, body, true)
	if first.Code != http.StatusOK {
		t.Fatalf("retry=%d", first.Code)
	}
	reopened, err := OpenStoreWithConfig(t.Context(), StoreConfig{SnapshotStore: snapshot})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(reopened.Close)
	f.handler = NewServer(ServerConfig{Assignment: reopened, AIAgentClient: f.devices, CLIInstallationService: f.service}).Handler()
	again := f.post(t, path, body, true)
	if again.Code != http.StatusOK || again.Body.String() != first.Body.String() {
		t.Fatal("restart lost consumed intent or immutable link ACK")
	}
	policy, err := reopened.readCLIInstallationPolicy(t.Context(), f.request.cliInstallationIdentity, f.devices)
	if err != nil || policy.Decision != "unknown" || policy.ExecutionRevision != "1" {
		t.Fatalf("restart became allowed or changed revision: %+v %v", policy, err)
	}
}

func TestCLIInstallationUnsupportedJournalFailsClosed(t *testing.T) {
	f := newCLIInstallationFixture(t, StoreConfig{OperationStore: &runtimeFakeAssignmentOperationStore{}})
	if w := f.post(t, "/v1/internal/cli-installations/link-intents", f.request, false); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("unsupported authority=%d", w.Code)
	}
}

func TestCLIInstallationRejectsTwoPersistenceAuthorities(t *testing.T) {
	_, newStore, _ := newTerminationDynamoFixture(t)
	snapshot, err := NewFileStoreSnapshot(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	f := newCLIInstallationFixture(t, StoreConfig{OperationStore: newStore(), SnapshotStore: snapshot})
	if w := f.post(t, "/v1/internal/cli-installations/link-intents", f.request, false); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("dual persistence authority accepted=%d", w.Code)
	}
}
