package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"testing"
)

type evidenceSnapshot struct {
	SnapshotStore
	saves     int
	loadError bool
}

func (s *evidenceSnapshot) SaveStoreSnapshot(ctx context.Context, snapshot StoreSnapshot) error {
	s.saves++
	return s.SnapshotStore.SaveStoreSnapshot(ctx, snapshot)
}

func (s *evidenceSnapshot) LoadStoreSnapshot(ctx context.Context) (StoreSnapshot, bool, error) {
	if s.loadError {
		return StoreSnapshot{}, false, errors.New("synthetic read failure")
	}
	return s.SnapshotStore.LoadStoreSnapshot(ctx)
}

func TestCLIInstallationIntentEvidenceRequiresPersistedSnapshot(t *testing.T) {
	for _, committed := range []bool{false, true} {
		t.Run(map[bool]string{false: "before-save", true: "after-save"}[committed], func(t *testing.T) {
			file, err := NewFileStoreSnapshot(filepath.Join(t.TempDir(), "snapshot.json"))
			if err != nil {
				t.Fatal(err)
			}
			uncertain := &uncertainIntentSnapshot{SnapshotStore: file}
			snapshot := &evidenceSnapshot{SnapshotStore: uncertain}
			f := newCLIInstallationFixture(t, StoreConfig{SnapshotStore: snapshot})
			body := installationIntentReplayBody(t, f)
			if got := f.post(t, installationEvidencePath, installationEvidenceBody(body), false); got.Code != http.StatusOK {
				t.Fatalf("empty status=%d", got.Code)
			}
			uncertain.failBeforeSave, uncertain.failAfterSave = !committed, committed
			if got := f.post(t, "/v1/internal/cli-installations/link-intents", body, false); got.Code != http.StatusServiceUnavailable {
				t.Fatal("missing synthetic write failure")
			}
			saves := snapshot.saves
			want := http.StatusServiceUnavailable
			if committed {
				want = http.StatusOK
			}
			if got := f.post(t, installationEvidencePath, installationEvidenceBody(body), false); got.Code != want {
				t.Fatalf("uncertain evidence status=%d want=%d", got.Code, want)
			}
			snapshot.loadError = true
			if got := f.post(t, installationEvidencePath, installationEvidenceBody(body), false); got.Code != http.StatusServiceUnavailable {
				t.Fatal("unreadable snapshot returned evidence")
			}
			if snapshot.saves != saves {
				t.Fatal("evidence read invoked snapshot save")
			}
			snapshot.loadError = false
			if committed {
				f.actor.Close()
				reopened, err := OpenStoreWithConfig(t.Context(), StoreConfig{SnapshotStore: snapshot})
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(reopened.Close)
				f.handler = NewServer(ServerConfig{Assignment: reopened, AIAgentClient: f.devices, CLIInstallationService: f.service}).Handler()
				saves = snapshot.saves // Store.Close persists independently of this read.
				if got := f.post(t, installationEvidencePath, installationEvidenceBody(body), false); got.Code != http.StatusOK {
					t.Fatal("persisted evidence lost after restart")
				}
				if snapshot.saves != saves {
					t.Fatal("restart evidence read invoked save")
				}
			}
		})
	}
}
