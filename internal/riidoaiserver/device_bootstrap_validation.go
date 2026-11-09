package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"reflect"
)

func (q deviceBootstrapRequest) valid() error {
	if q.SchemaVersion != cliInstallationIntentSchema || !installationRandom32(q.TransitionID) || !bootstrapDigest(q.NextSecretHash) || !bootstrapLegacyDevice(q.ExpectedDeviceID) {
		return errCLIInstallationInvalid
	}
	_, err := installationPublicKey(q.cliInstallationIntentRequest)
	return err
}

func (r deviceBootstrapRecord) valid() error {
	q, b, a := r.Request, r.Before, r.After
	if q.valid() != nil || r.Schema != deviceBootstrapSchema || r.Revision == 0 || b.DeviceID != q.ExpectedDeviceID || b.OwnerPrincipalID != q.ApprovedPrincipalID || !cliInstallationID(b.WorkspaceID) || !cliInstallationID(b.MachineID) || !bootstrapDigest(b.SecretHashSHA256) || b.SecretHashSHA256 == q.NextSecretHash || b.IssuedAt.IsZero() || b.ExecutionCoverageGeneration != "" || b.CredentialGeneration != "" {
		return errCLIInstallationInvalid
	}
	if a.SecretHashSHA256 != q.NextSecretHash || a.ExecutionCoverageGeneration != bootstrapGeneration(q) || !cliInstallationID(a.CredentialGeneration) || a.CredentialGeneration == "" || a.IssuedAt.IsZero() {
		return errCLIInstallationInvalid
	}
	// Identity, ownership, workspace, label and machine are immutable across the
	// transition. No caller-controlled replacement device is accepted.
	a.SecretHashSHA256, a.ExecutionCoverageGeneration, a.CredentialGeneration, a.IssuedAt = b.SecretHashSHA256, "", "", b.IssuedAt
	if !reflect.DeepEqual(a, b) || r.Intent.Request != q.cliInstallationIntentRequest || !installationRandom32(r.Intent.ID) || !bootstrapDigest(r.Intent.ChallengeDigest) || r.Intent.ExpiresAt.IsZero() || r.Intent.Consumed || r.Intent.Retired || r.Intent.ReplayKey != "" {
		return errCLIInstallationInvalid
	}
	switch r.State {
	case "issued":
		if r.PolicyRevision != 0 || r.PreviousBootDigest != "" || r.CompletedBootDigest != "" || !r.CompletedAt.IsZero() {
			return errCLIInstallationInvalid
		}
	case "prepared":
		if r.PolicyRevision == 0 || r.PolicyRevision == math.MaxUint64 || !bootstrapDigest(r.PreviousBootDigest) || r.CompletedBootDigest != "" || !r.CompletedAt.IsZero() {
			return errCLIInstallationInvalid
		}
	case "complete":
		if r.PolicyRevision < 2 || !bootstrapDigest(r.PreviousBootDigest) || !bootstrapDigest(r.CompletedBootDigest) || r.PreviousBootDigest == r.CompletedBootDigest || r.CompletedAt.IsZero() {
			return errCLIInstallationInvalid
		}
	default:
		return errCLIInstallationInvalid
	}
	return nil
}

func bootstrapGeneration(q deviceBootstrapRequest) string {
	digest := sha256.Sum256([]byte(deviceBootstrapSchema + "\x00" + q.TransitionID + "\x00" + q.ExpectedDeviceID + "\x00" + q.KeyFingerprint + "\x00" + q.NextSecretHash))
	return deviceBootstrapCoveragePrefix + hex.EncodeToString(digest[:])
}
