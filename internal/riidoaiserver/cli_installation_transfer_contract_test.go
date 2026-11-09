package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCLIInstallationTransferRequiresSeparateScope(t *testing.T) {
	f := newCLITransferFixture(t)
	// Restore the pre-transfer service configuration: link/policy-read only.
	f.cli.handler = NewServer(ServerConfig{Assignment: f.cli.actor, AIAgentClient: f.devices, Authorizer: f.authorizer, CLIInstallationService: f.cli.service}).Handler()
	w := f.post(t, "/v1/internal/cli-installations/link-intents", f.wire, "service", "")
	if w.Code != http.StatusForbidden {
		t.Fatalf("ordinary link authority accepted transfer: %d", w.Code)
	}
	f.assertOriginal(t)
}

func TestCLIInstallationTransferRejectsInvalidAndLegacyFields(t *testing.T) {
	f := newCLITransferFixture(t)
	for _, revision := range []string{"", "0", "01", "+1", "-1", "18446744073709551615", "18446744073709551616"} {
		t.Run("revision="+revision, func(t *testing.T) {
			wire := f.wire
			wire.PreviousLinkRevision = revision
			if w := f.post(t, "/v1/internal/cli-installations/link-intents", wire, "service", ""); w.Code != http.StatusBadRequest {
				t.Fatalf("malformed revision accepted=%d", w.Code)
			}
		})
	}
	for _, schema := range []string{cliInstallationIntentSchema, cliInstallationIntentV2Schema} {
		wire := f.wire
		wire.SchemaVersion = schema
		if w := f.post(t, "/v1/internal/cli-installations/link-intents", wire, "service", ""); w.Code != http.StatusBadRequest {
			t.Fatalf("legacy schema accepted transfer fields=%d", w.Code)
		}
	}
	wire := f.wire
	wire.PreviousDeviceID = wire.ExpectedDeviceID
	if w := f.post(t, "/v1/internal/cli-installations/link-intents", wire, "service", ""); w.Code != http.StatusBadRequest {
		t.Fatalf("same-device transfer accepted=%d", w.Code)
	}
	f.assertOriginal(t)
}

func TestCLIInstallationTransferReplayBindsImmutableSource(t *testing.T) {
	f := newCLITransferFixture(t)
	wire := f.wire
	wire.PreviousLinkRevision = "2"
	if w := f.post(t, "/v1/internal/cli-installations/link-intents", wire, "service", ""); w.Code != http.StatusConflict {
		t.Fatalf("changed source replay=%d", w.Code)
	}
	w := f.post(t, "/v1/internal/cli-installations/link-intents", f.wire, "service", "")
	var replay cliInstallationIntentReplayResponse
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &replay) != nil || replay != f.intent {
		t.Fatal("original intent not preserved")
	}
	f.assertOriginal(t)
}
