package riidoaiserver

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
)

type bootstrapFixture struct {
	dynamo     *membershipSharedDynamoFixture
	store      *PersistentAIAgentClientStore
	key        *ecdsa.PrivateKey
	request    deviceBootstrapRequest
	credential EnrollDeviceResponse
	nextSecret string
}

func newBootstrapFixture(t *testing.T) *bootstrapFixture {
	t.Helper()
	f := &bootstrapFixture{dynamo: newMembershipSharedDynamoFixture(t), nextSecret: "rdev_" + strings.Repeat("2", 64)}
	f.store = openCoverageEnrollmentFixture(t, f.dynamo.server.URL, false)
	principal := profileColorTestPrincipal()
	var err error
	f.credential, err = f.store.EnrollDeviceCredential(context.Background(), principal, principal.WorkspaceID, EnrollDeviceRequest{MachineID: "bootstrap-machine"})
	if err != nil {
		t.Fatal(err)
	}
	f.key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	public, err := f.key.PublicKey.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	fingerprint, nextHash := sha256.Sum256(public), sha256.Sum256([]byte(f.nextSecret))
	f.request = deviceBootstrapRequest{TransitionID: base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("t", 32))), NextSecretHash: hex.EncodeToString(nextHash[:]), cliInstallationIntentRequest: cliInstallationIntentRequest{SchemaVersion: cliInstallationIntentSchema, cliInstallationIdentity: cliInstallationIdentity{Resource: "https://cli.fixture.invalid", ClientID: "riido-cli", InstallationID: "bootstrap-install", KeyFingerprint: hex.EncodeToString(fingerprint[:])}, PublicKeyAlgorithm: cliInstallationAlgorithm, PublicKey: base64.RawURLEncoding.EncodeToString(public), ExpectedDeviceID: f.credential.DeviceID, ApprovedPrincipalID: principal.PrincipalID}}
	if err := f.request.valid(); err != nil {
		t.Fatalf("fixture request invalid: device=%q owner=%q err=%v", f.request.ExpectedDeviceID, f.request.ApprovedPrincipalID, err)
	}
	return f
}

func (f *bootstrapFixture) auth() deviceBootstrapAuth {
	return deviceBootstrapAuth{deviceID: f.credential.DeviceID, ownerPrincipalID: f.credential.OwnerPrincipalID}
}

func bootstrapAuthFor(c AIAgentClientDeviceCredentialSnapshot) deviceBootstrapAuth {
	return deviceBootstrapAuth{deviceID: c.DeviceID, ownerPrincipalID: c.OwnerPrincipalID, executionCoverageGeneration: c.ExecutionCoverageGeneration, credentialGeneration: c.CredentialGeneration}
}
