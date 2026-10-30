package riidoaiserver

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCLIInstallationRetirementCannotPublishOversizedSnapshot(t *testing.T) {
	f := newCLIInstallationFixture(t)
	response := f.intent(t)
	intent := cliInstallationIntent{Request: f.request, ID: response.IntentID, ChallengeDigest: strings.Repeat("0", 64), ExpiresAt: time.Now().UTC().Add(-time.Hour)}
	// These opaque identity fields accept escaped JSON characters. Construct
	// an accepted record just below the bound, then exercise the transition.
	intent.Request.InstallationID = strings.Repeat("\"", 512)
	intent.Request.ExpectedDeviceID = strings.Repeat("\"", 512)
	for n := 1; n <= 512; n++ {
		intent.Request.ApprovedPrincipalID = strings.Repeat("a", n)
		raw, err := json.Marshal(intent)
		if err != nil {
			t.Fatal(err)
		}
		if len(raw) > (3<<10)-14 {
			break
		}
	}
	if err := intent.valid(); err != nil {
		t.Fatalf("original record invalid: %v", err)
	}
	retired := intent
	retired.Retired = true
	if retired.valid() == nil {
		t.Fatal("fixture must cross the bound on retirement")
	}
	state := storeState{cliInstallationIntents: map[string]cliInstallationIntent{intent.ID: intent}}
	req := cliInstallationRetirementRequest{SchemaVersion: cliInstallationRetirementSchema, IntentID: intent.ID, cliInstallationIdentity: intent.Request.cliInstallationIdentity, ExpectedDeviceID: intent.Request.ExpectedDeviceID, ApprovedPrincipalID: intent.Request.ApprovedPrincipalID}
	got := f.actor.commitCLIInstallationRetirement(&state, t.Context(), req)
	if !errors.Is(got.err, errCLIInstallationInvalid) || state.cliInstallationIntents[intent.ID] != intent {
		t.Fatal("oversized retired record published into snapshot state")
	}
}
