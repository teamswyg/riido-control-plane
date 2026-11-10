package riidoaiserver

import (
	"errors"
	"testing"
)

func TestCLIInstallationTransferEvidenceSnapshotSeparatesPurposeAndCorruption(t *testing.T) {
	f := newCLITransferFixture(t)
	wire := f.wire
	wire.SchemaVersion = cliInstallationTransferEvidenceSchema
	lookup, err := installationEvidenceLookup(wire)
	if err != nil {
		t.Fatal(err)
	}
	record := cliInstallationIntentReplay{f.wire.cliInstallationIntentRequest, f.intent}
	for _, state := range []string{"valid", "request-only", "intent-only", "challenge", "source"} {
		t.Run(state, func(t *testing.T) {
			snapshot := cliInstallationSnapshot{
				IntentReplays: map[string]cliInstallationIntentReplay{record.key(): record},
				Intents:       map[string]cliInstallationIntent{record.Response.IntentID: record.intent()},
			}
			switch state {
			case "request-only":
				delete(snapshot.Intents, record.Response.IntentID)
			case "intent-only":
				delete(snapshot.IntentReplays, record.key())
			case "challenge", "source":
				broken := record.intent()
				if state == "source" {
					broken.Request.PreviousLinkRevision = "2"
				} else {
					broken.ChallengeDigest = "broken"
				}
				snapshot.Intents[record.Response.IntentID] = broken
			}
			got, err := readInstallationEvidence(snapshot, lookup)
			if state == "valid" {
				if err != nil || !got.Found || got.Intent == nil || *got.Intent != f.intent {
					t.Fatal("snapshot lost immutable transfer evidence")
				}
			} else if !errors.Is(err, errCLIInstallationUnavailable) {
				t.Fatal("corrupt snapshot accepted")
			}
		})
	}
	initial := wire
	initial.SchemaVersion = cliInstallationIntentEvidenceSchema
	initial.PreviousDeviceID, initial.PreviousLinkRevision = "", ""
	other, err := installationEvidenceLookup(initial)
	if err != nil || other.key() == lookup.key() || other.intentID() == lookup.intentID() {
		t.Fatal("initial link and transfer namespaces collided")
	}
	snapshot := cliInstallationSnapshot{IntentReplays: map[string]cliInstallationIntentReplay{record.key(): record}, Intents: map[string]cliInstallationIntent{record.Response.IntentID: record.intent()}}
	got, err := readInstallationEvidence(snapshot, other)
	if err != nil || got.Found || got.Intent != nil || got.SchemaVersion != cliInstallationIntentEvidenceSchema {
		t.Fatal("initial evidence returned transfer authority")
	}
}
