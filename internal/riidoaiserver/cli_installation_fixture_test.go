package riidoaiserver

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"testing"
)

type cliInstallationFixture struct {
	actor   *Store
	devices *DevelopmentAIAgentClientStore
	handler http.Handler
	key     *ecdsa.PrivateKey
	request cliInstallationIntentRequest
	service CLIInstallationServiceConfig
}

func newCLIInstallationFixture(t *testing.T, configs ...StoreConfig) *cliInstallationFixture {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pub, err := key.PublicKey.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(pub)
	resource := "https://cli.fixture.invalid"
	authority, err := NewCLIInstallationServiceConfig(resource, []StaticTokenCredential{{PrincipalID: cliInstallationServicePrincipal, Token: "service-fixture-token", Scopes: []string{"cli-installation:link", "cli-installation:policy-read"}}})
	if err != nil {
		t.Fatal(err)
	}
	actor := NewStore()
	if len(configs) > 0 {
		actor.Close()
		actor = NewStoreWithConfig(configs[0])
	}
	t.Cleanup(actor.Close)
	devices := NewDevelopmentAIAgentClientStore()
	devices.deviceCredentials["fixture-device"] = deviceCredentialRecord{deviceID: "fixture-device", ownerPrincipalID: "owner", secretHash: sha256.Sum256([]byte("fixture-device-secret"))}
	req := cliInstallationIntentRequest{SchemaVersion: cliInstallationIntentSchema, cliInstallationIdentity: cliInstallationIdentity{"installation", resource, "riido-cli", hex.EncodeToString(digest[:])}, PublicKeyAlgorithm: cliInstallationAlgorithm, PublicKey: base64.RawURLEncoding.EncodeToString(pub), ExpectedDeviceID: "fixture-device", ApprovedPrincipalID: "owner"}
	return &cliInstallationFixture{actor, devices, NewServer(ServerConfig{Assignment: actor, AIAgentClient: devices, CLIInstallationService: authority}).Handler(), key, req, authority}
}

func (f *cliInstallationFixture) signature(t *testing.T, response cliInstallationIntentResponse) string {
	t.Helper()
	challenge, err := base64.RawURLEncoding.DecodeString(response.Challenge)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(challenge)
	intent := cliInstallationIntent{Request: f.request, ID: response.IntentID, ChallengeDigest: hex.EncodeToString(digest[:]), ExpiresAt: response.ExpiresAt}
	hash := sha256.Sum256(cliInstallationLinkMessage(intent))
	sig, err := ecdsa.SignASN1(rand.Reader, f.key, hash[:])
	if err != nil {
		t.Fatal(err)
	}
	return base64.RawURLEncoding.EncodeToString(sig)
}
