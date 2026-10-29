package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestCLIInstallationReceiptSurvivesExpiryAndSnapshotRestart(t *testing.T) {
	var clock atomic.Int64
	clock.Store(time.Now().UTC().UnixNano())
	now := func() time.Time { return time.Unix(0, clock.Load()).UTC() }
	snapshot, err := NewFileStoreSnapshot(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := StoreConfig{Now: now, SnapshotStore: snapshot}
	f := newCLIInstallationFixture(t, cfg)
	intent := f.intent(t)
	unconsumed := f.intent(t)
	confirmed := f.post(t, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", map[string]string{
		"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent),
	}, true)
	if confirmed.Code != http.StatusOK {
		t.Fatalf("confirm status=%d", confirmed.Code)
	}
	clock.Store(intent.ExpiresAt.Add(time.Hour).UnixNano())
	reopened, err := OpenStoreWithConfig(t.Context(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(reopened.Close)
	f.handler = NewServer(ServerConfig{Assignment: reopened, AIAgentClient: f.devices, CLIInstallationService: f.service}).Handler()
	path := "/v1/internal/cli-installations/link-receipts"
	got := f.post(t, path, installationReceiptBody(f, intent.IntentID), false)
	var receipt cliInstallationReceipt
	var original cliInstallationLink
	if got.Code != http.StatusOK || json.Unmarshal(got.Body.Bytes(), &receipt) != nil || json.Unmarshal(confirmed.Body.Bytes(), &original) != nil || receipt.Link != original {
		t.Fatalf("expired consumed receipt lost original ACK: status=%d", got.Code)
	}
	if got := f.post(t, path, installationReceiptBody(f, unconsumed.IntentID), false); got.Code != http.StatusGone {
		t.Fatalf("expired unconsumed intent became receipt: status=%d", got.Code)
	}
	policy, err := reopened.readCLIInstallationPolicy(t.Context(), f.request.cliInstallationIdentity, f.devices)
	if err != nil || policy.Decision != "unknown" || policy.ExecutionRevision != "1" {
		t.Fatalf("historical receipt changed permission or revision: %+v %v", policy, err)
	}
}
