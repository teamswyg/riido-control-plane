package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"
)

func TestCLIInstallationCanonicalProofBindsDomainAndEveryIdentityField(t *testing.T) {
	expiry, _ := time.Parse(time.RFC3339Nano, "2026-10-28T01:02:03.123Z")
	intent := cliInstallationIntent{ID: "intent", ChallengeDigest: strings.Repeat("a", 64), ExpiresAt: expiry, Request: cliInstallationIntentRequest{cliInstallationIdentity: cliInstallationIdentity{InstallationID: "installation", Resource: "https://cli.fixture.invalid", ClientID: "riido-cli", KeyFingerprint: strings.Repeat("b", 64)}, ExpectedDeviceID: "device", ApprovedPrincipalID: "approved-human"}}
	digest := sha256.Sum256(cliInstallationLinkMessage(intent))
	// Independently encoded uint32 big-endian length-prefixed UTF-8 fixture.
	if hex.EncodeToString(digest[:]) != "979957be1deb56d1c5f4a68a6153ab690cc2c394d2551c3b0f1d7ed99db3a809" {
		t.Fatal("canonical signature wire changed")
	}
	f := newCLIInstallationFixture(t)
	response := f.intent(t)
	challenge := sha256.Sum256(mustDecodeCLIChallenge(t, response.Challenge))
	valid := cliInstallationIntent{ID: response.IntentID, Request: f.request, ChallengeDigest: hex.EncodeToString(challenge[:]), ExpiresAt: response.ExpiresAt}
	sig := f.signature(t, response)
	for _, field := range []string{"resource", "client", "installation", "intent", "challenge", "fingerprint", "device", "approver", "expiry"} {
		changed := valid
		switch field {
		case "resource":
			changed.Request.Resource = "https://other.fixture.invalid"
		case "client":
			changed.Request.ClientID = "another-client"
		case "installation":
			changed.Request.InstallationID += "-other"
		case "intent":
			changed.ID += "-other"
		case "challenge":
			changed.ChallengeDigest = strings.Repeat("0", 64)
		case "fingerprint":
			changed.Request.KeyFingerprint = strings.Repeat("0", 64)
		case "device":
			changed.Request.ExpectedDeviceID = "another-device"
		case "approver":
			changed.Request.ApprovedPrincipalID = "another-owner"
		case "expiry":
			changed.ExpiresAt = changed.ExpiresAt.Add(time.Second)
		}
		if err := verifyCLIInstallationProof(changed, authenticatedClaimDevice{deviceID: "fixture-device", ownerPrincipalID: "owner"}, sig); err == nil {
			t.Fatalf("unsigned %s substitution accepted", field)
		}
	}
}
