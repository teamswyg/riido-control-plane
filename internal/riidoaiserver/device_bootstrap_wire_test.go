package riidoaiserver

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
	"time"
)

func TestDeviceBootstrapWireMatchesProtectedNativeHelper(t *testing.T) {
	expires, err := time.Parse(time.RFC3339Nano, "2026-10-28T01:02:03.123Z")
	if err != nil {
		t.Fatal(err)
	}
	r := deviceBootstrapRecord{Request: deviceBootstrapRequest{TransitionID: base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("t", 32))), cliInstallationIntentRequest: cliInstallationIntentRequest{cliInstallationIdentity: cliInstallationIdentity{Resource: "https://cli.fixture.invalid", ClientID: "riido-cli", InstallationID: "installation", KeyFingerprint: strings.Repeat("b", 64)}, ExpectedDeviceID: "device", ApprovedPrincipalID: "approved-human"}}, Intent: cliInstallationIntent{ID: base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("i", 32))), ChallengeDigest: strings.Repeat("a", 64), ExpiresAt: expires}}
	digest := sha256.Sum256(deviceBootstrapMessage(r, "prepare", strings.Repeat("c", 64)))
	if hex.EncodeToString(digest[:]) != "3cc36ae5f4cf0c3264b303853809ae2d0085eb466e3eedda70651ca17c06d009" {
		t.Fatal("server/native bootstrap wire contract differs")
	}
}
