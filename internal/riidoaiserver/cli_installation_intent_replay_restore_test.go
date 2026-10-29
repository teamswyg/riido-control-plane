package riidoaiserver

import (
	"encoding/base64"
	"path/filepath"
	"testing"
)

func TestCLIInstallationIntentV2SnapshotRejectsOrphanPairs(t *testing.T) {
	for _, broken := range []string{"request-only", "intent-only", "challenge"} {
		t.Run(broken, func(t *testing.T) {
			file, err := NewFileStoreSnapshot(filepath.Join(t.TempDir(), "snapshot.json"))
			if err != nil {
				t.Fatal(err)
			}
			f := newCLIInstallationFixture(t, StoreConfig{SnapshotStore: file})
			body := installationIntentReplayBody(t, f)
			first := f.post(t, "/v1/internal/cli-installations/link-intents", body, false)
			response := assertInstallationIntentReplayResponse(t, first, body["request_id"])
			f.actor.Close()
			snapshot, found, err := file.LoadStoreSnapshot(t.Context())
			if err != nil || !found {
				t.Fatalf("original snapshot missing: %v", err)
			}
			switch broken {
			case "request-only":
				delete(snapshot.CLIInstallations.Intents, response.IntentID)
			case "intent-only":
				snapshot.CLIInstallations.IntentReplays = nil
			case "challenge":
				for key, record := range snapshot.CLIInstallations.IntentReplays {
					record.Response.Challenge = base64.RawURLEncoding.EncodeToString(make([]byte, 32))
					snapshot.CLIInstallations.IntentReplays[key] = record
				}
			}
			if err := file.SaveStoreSnapshot(t.Context(), snapshot); err != nil {
				t.Fatal(err)
			}
			restored, err := OpenStoreWithConfig(t.Context(), StoreConfig{SnapshotStore: file})
			if err == nil {
				restored.Close()
				t.Fatal("incomplete v2 snapshot restored as legacy authority")
			}
		})
	}
}
