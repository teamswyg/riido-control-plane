package riidoaiserver

import (
	"net/http"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestCLIInstallationRetirementLostSnapshotACKCannotEraseFence(t *testing.T) {
	var clock atomic.Int64
	clock.Store(time.Now().UTC().UnixNano())
	now := func() time.Time { return time.Unix(0, clock.Load()).UTC() }
	file, err := NewFileStoreSnapshot(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	snapshot := &uncertainIntentSnapshot{SnapshotStore: file}
	f := newCLIInstallationFixture(t, StoreConfig{Now: now, SnapshotStore: snapshot})
	intent := f.intent(t)
	clock.Store(intent.ExpiresAt.UnixNano())
	body := installationRetireBody(f, intent.IntentID)
	snapshot.failAfterSave = true
	if got := f.post(t, installationRetirePath, body, false); got.Code != http.StatusServiceUnavailable {
		t.Fatalf("uncertain retirement acknowledged: status=%d", got.Code)
	}
	snapshot.failAfterSave = false
	_ = f.intent(t) // An unrelated snapshot-producing command must retain the fence.
	saved, found, err := file.LoadStoreSnapshot(t.Context())
	if err != nil || !found || !saved.CLIInstallations.Intents[intent.IntentID].Retired {
		t.Fatal("later save erased uncertain retirement")
	}
	snapshot.failBeforeSave = true
	if got := f.post(t, installationRetirePath, body, false); got.Code != http.StatusServiceUnavailable {
		t.Fatalf("retirement acknowledged without successful persistence: status=%d", got.Code)
	}
	snapshot.failBeforeSave = false
	if got := f.post(t, installationRetirePath, body, false); got.Code != http.StatusOK {
		t.Fatalf("retry did not acknowledge fence: status=%d", got.Code)
	}
	corrupt := saved.CLIInstallations.Intents[intent.IntentID]
	corrupt.Consumed = true
	saved.CLIInstallations.Intents[intent.IntentID] = corrupt
	if err := file.SaveStoreSnapshot(t.Context(), saved); err != nil {
		t.Fatal(err)
	}
	if reopened, err := OpenStoreWithConfig(t.Context(), StoreConfig{Now: now, SnapshotStore: file}); err == nil {
		reopened.Close()
		t.Fatal("contradictory consumed+retired snapshot accepted")
	}
}
