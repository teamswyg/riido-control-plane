package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

const installationRetirePath = "/v1/internal/cli-installations/link-intent-retirements"

func installationRetireBody(f *cliInstallationFixture, id string) map[string]string {
	body := installationReceiptBody(f, id)
	body["schema_version"] = "riido-cli-installation-link-intent-retirement.v1"
	return body
}

func TestCLIInstallationRetirementRequiresExpiryAndSurvivesRestart(t *testing.T) {
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
	body := installationRetireBody(f, intent.IntentID)
	if got := f.post(t, installationRetirePath, body, false); got.Code != http.StatusConflict {
		t.Fatalf("live intent retirement status=%d", got.Code)
	}
	clock.Store(intent.ExpiresAt.UnixNano())
	first := f.post(t, installationRetirePath, body, false)
	var result struct {
		Status        string `json:"status"`
		IntentID      string `json:"intent_id"`
		SchemaVersion string `json:"schema_version"`
	}
	if first.Code != http.StatusOK || json.Unmarshal(first.Body.Bytes(), &result) != nil || result.Status != "retired" || result.IntentID != intent.IntentID || result.SchemaVersion != body["schema_version"] {
		t.Fatalf("expired intent retirement status=%d", first.Code)
	}
	f.actor.Close()
	reopened, err := OpenStoreWithConfig(t.Context(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(reopened.Close)
	f.handler = NewServer(ServerConfig{Assignment: reopened, AIAgentClient: f.devices, CLIInstallationService: f.service}).Handler()
	if got := f.post(t, installationRetirePath, body, false); got.Code != http.StatusOK || got.Body.String() != first.Body.String() {
		t.Fatalf("retirement receipt changed on restart: status=%d", got.Code)
	}
	// Even a clock moving back cannot reopen a durably retired intent.
	clock.Store(intent.ExpiresAt.Add(-time.Second).UnixNano())
	if got := f.post(t, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent)}, true); got.Code != http.StatusGone {
		t.Fatalf("retired intent confirmed: status=%d", got.Code)
	}
}
