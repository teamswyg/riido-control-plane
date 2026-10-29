package riidoaiserver

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"time"
)

func cliInstallationLinkMessage(intent cliInstallationIntent) []byte {
	r := intent.Request
	fields := []string{"riido-cli-device-link.v1", "link", r.Resource, r.ClientID, r.InstallationID, intent.ID, intent.ChallengeDigest, r.KeyFingerprint, r.ExpectedDeviceID, r.ApprovedPrincipalID, intent.ExpiresAt.UTC().Format(time.RFC3339Nano)}
	var message bytes.Buffer
	for _, field := range fields {
		_ = binary.Write(&message, binary.BigEndian, uint32(len(field)))
		message.WriteString(field)
	}
	return message.Bytes()
}

func verifyCLIInstallationProof(intent cliInstallationIntent, device authenticatedClaimDevice, signature string) error {
	key, err := installationPublicKey(intent.Request)
	if err != nil || device.deviceID != intent.Request.ExpectedDeviceID || device.ownerPrincipalID != intent.Request.ApprovedPrincipalID {
		return ErrAuthorizationForbidden
	}
	proof, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil || len(proof) > 80 || base64.RawURLEncoding.EncodeToString(proof) != signature {
		return ErrAuthorizationForbidden
	}
	digest := sha256.Sum256(cliInstallationLinkMessage(intent))
	if !ecdsa.VerifyASN1(key, digest[:], proof) {
		return ErrAuthorizationForbidden
	}
	return nil
}

func linkForCLIInstallation(intent cliInstallationIntent, device authenticatedClaimDevice, now time.Time) cliInstallationLink {
	return cliInstallationLink{SchemaVersion: cliInstallationLinkSchema, cliInstallationIdentity: intent.Request.cliInstallationIdentity, DeviceID: device.deviceID, LinkApprovedBy: intent.Request.ApprovedPrincipalID, DeviceOwnerAtLink: device.ownerPrincipalID, LinkRevision: "1", LinkedAt: now.UTC()}
}

func matchingCLIInstallationLink(link cliInstallationLink, intent cliInstallationIntent, device authenticatedClaimDevice) bool {
	return link.SchemaVersion == cliInstallationLinkSchema && link.cliInstallationIdentity == intent.Request.cliInstallationIdentity && link.DeviceID == device.deviceID && link.DeviceOwnerAtLink == device.ownerPrincipalID && link.LinkApprovedBy == intent.Request.ApprovedPrincipalID && link.LinkRevision == "1" && !link.LinkedAt.IsZero()
}
