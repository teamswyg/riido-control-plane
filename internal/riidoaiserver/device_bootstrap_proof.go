package riidoaiserver

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"time"
)

func deviceBootstrapMessage(r deviceBootstrapRecord, phase, boot string) []byte {
	q, i := r.Request, r.Intent
	fields := []string{"riido-cli-device-bootstrap.v1", phase, q.Resource, q.ClientID, q.InstallationID, q.TransitionID, i.ID, i.ChallengeDigest, q.KeyFingerprint, q.ExpectedDeviceID, q.ApprovedPrincipalID, i.ExpiresAt.UTC().Format(time.RFC3339Nano), r.PreviousBootDigest, boot}
	var message bytes.Buffer
	for _, field := range fields {
		_ = binary.Write(&message, binary.BigEndian, uint32(len(field)))
		message.WriteString(field)
	}
	return message.Bytes()
}

func verifyDeviceBootstrapProof(r deviceBootstrapRecord, phase string, proof deviceBootstrapProof, now time.Time) error {
	if (phase != "prepare" && phase != "activate") || proof.IntentID != r.Intent.ID || !bootstrapDigest(proof.BootDigest) {
		return ErrAuthorizationForbidden
	}
	if !now.Before(r.Intent.ExpiresAt) {
		return errCLIInstallationExpired
	}
	if phase == "prepare" && (r.State != "issued" || r.PreviousBootDigest != "") {
		return errCLIInstallationConflict
	}
	if phase == "activate" && (r.State != "prepared" || !bootstrapDigest(r.PreviousBootDigest) || r.PreviousBootDigest == proof.BootDigest) {
		return errCLIInstallationConflict
	}
	key, err := installationPublicKey(r.Request.cliInstallationIntentRequest)
	if err != nil {
		return ErrAuthorizationForbidden
	}
	signature, err := base64.RawURLEncoding.DecodeString(proof.Signature)
	if err != nil || len(signature) > 80 || base64.RawURLEncoding.EncodeToString(signature) != proof.Signature {
		return ErrAuthorizationForbidden
	}
	digest := sha256.Sum256(deviceBootstrapMessage(r, phase, proof.BootDigest))
	if !ecdsa.VerifyASN1(key, digest[:], signature) {
		return ErrAuthorizationForbidden
	}
	return nil
}
