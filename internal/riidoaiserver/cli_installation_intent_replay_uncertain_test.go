package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"testing"
)

type uncertainIntentSnapshot struct {
	SnapshotStore
	failAfterSave  bool
	failBeforeSave bool
}

func (s *uncertainIntentSnapshot) SaveStoreSnapshot(ctx context.Context, snapshot StoreSnapshot) error {
	if s.failBeforeSave {
		return errors.New("synthetic unavailable snapshot")
	}
	if err := s.SnapshotStore.SaveStoreSnapshot(ctx, snapshot); err != nil {
		return err
	}
	if s.failAfterSave {
		return errors.New("synthetic lost snapshot acknowledgement")
	}
	return nil
}

func TestCLIInstallationIntentV2LostSnapshotACKPreservesOriginalNonce(t *testing.T) {
	file, err := NewFileStoreSnapshot(filepath.Join(t.TempDir(), "snapshot.json"))
	if err != nil {
		t.Fatal(err)
	}
	snapshots := &uncertainIntentSnapshot{SnapshotStore: file}
	f := newCLIInstallationFixture(t, StoreConfig{SnapshotStore: snapshots})
	body := installationIntentReplayBody(t, f)
	snapshots.failAfterSave = true
	first := f.post(t, "/v1/internal/cli-installations/link-intents", body, false)
	if first.Code != http.StatusServiceUnavailable {
		t.Fatalf("lost snapshot ACK status=%d", first.Code)
	}
	saved, found, err := file.LoadStoreSnapshot(t.Context())
	if err != nil || !found {
		t.Fatal("synthetic committed snapshot missing")
	}
	var original cliInstallationIntentReplay
	for _, value := range saved.CLIInstallations.IntentReplays {
		original = value
	}
	if original.valid() != nil {
		t.Fatal("original committed pair missing")
	}
	snapshots.failAfterSave = false
	// Another actor command must not rewrite a snapshot from pre-error memory
	// and erase the original durable pair before this request retries.
	_ = f.intent(t)
	afterOtherSave, found, err := file.LoadStoreSnapshot(t.Context())
	if err != nil || !found || afterOtherSave.CLIInstallations.IntentReplays[original.key()] != original {
		t.Fatal("another snapshot save erased the uncertain original intent pair")
	}
	snapshots.failBeforeSave = true
	if got := f.post(t, "/v1/internal/cli-installations/link-intents", body, false); got.Code != http.StatusServiceUnavailable {
		t.Fatal("uncertain intent replay acknowledged without successful persistence")
	}
	snapshots.failBeforeSave = false
	again := f.post(t, "/v1/internal/cli-installations/link-intents", body, false)
	response := assertInstallationIntentReplayResponse(t, again, body["request_id"])
	if response.IntentID != original.Response.IntentID || response.Challenge != original.Response.Challenge || !response.ExpiresAt.Equal(original.Response.ExpiresAt) {
		t.Fatal("retry changed the original durable intent nonce or expiry")
	}
	f.actor.Close()
	reopened, err := OpenStoreWithConfig(t.Context(), StoreConfig{SnapshotStore: file})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(reopened.Close)
	f.handler = NewServer(ServerConfig{Assignment: reopened, AIAgentClient: f.devices, CLIInstallationService: f.service}).Handler()
	restart := f.post(t, "/v1/internal/cli-installations/link-intents", body, false)
	if restart.Code != http.StatusOK || restart.Body.String() != again.Body.String() {
		t.Fatal("restart changed recovered response")
	}
}
