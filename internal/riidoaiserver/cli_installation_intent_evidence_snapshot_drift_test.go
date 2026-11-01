package riidoaiserver

import (
	"net/http"
	"path/filepath"
	"testing"
	"time"
)

func TestCLIInstallationIntentEvidenceRejectsSnapshotDrift(t *testing.T) {
	for _, broken := range []string{"missing-pair", "intent-only", "request-only", "changed-pair", "schema"} {
		t.Run(broken, func(t *testing.T) {
			file, err := NewFileStoreSnapshot(filepath.Join(t.TempDir(), "snapshot.json"))
			if err != nil {
				t.Fatal(err)
			}
			snapshot := &evidenceSnapshot{SnapshotStore: file}
			f := newCLIInstallationFixture(t, StoreConfig{SnapshotStore: snapshot})
			body := installationIntentReplayBody(t, f)
			original := assertInstallationIntentReplayResponse(t, f.post(t, "/v1/internal/cli-installations/link-intents", body, false), body["request_id"])
			saved, found, err := file.LoadStoreSnapshot(t.Context())
			if err != nil || !found {
				t.Fatal("fixture snapshot missing")
			}
			for key, record := range saved.CLIInstallations.IntentReplays {
				switch broken {
				case "schema":
					saved.SchemaVersion = "unsupported.fixture"
				case "missing-pair":
					delete(saved.CLIInstallations.IntentReplays, key)
					delete(saved.CLIInstallations.Intents, original.IntentID)
				case "intent-only":
					delete(saved.CLIInstallations.IntentReplays, key)
				case "request-only":
					delete(saved.CLIInstallations.Intents, original.IntentID)
				case "changed-pair":
					record.Response.ExpiresAt = record.Response.ExpiresAt.Add(time.Second)
					saved.CLIInstallations.IntentReplays[key] = record
					saved.CLIInstallations.Intents[original.IntentID] = record.intent()
				}
			}
			if err := file.SaveStoreSnapshot(t.Context(), saved); err != nil {
				t.Fatal(err)
			}
			saves := snapshot.saves
			if got := f.post(t, installationEvidencePath, installationEvidenceBody(body), false); got.Code != http.StatusServiceUnavailable {
				t.Fatalf("drift %s status=%d", broken, got.Code)
			}
			if snapshot.saves != saves {
				t.Fatal("drift read wrote a replacement snapshot")
			}
		})
	}
}
