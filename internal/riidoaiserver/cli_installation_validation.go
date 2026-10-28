package riidoaiserver

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

func cliInstallationID(value string) bool {
	return value != "" && len(value) <= 512 && strings.TrimSpace(value) == value && !strings.ContainsAny(value, "/\\\x00\r\n\t")
}

func (id cliInstallationIdentity) valid() bool {
	digest, err := hex.DecodeString(id.KeyFingerprint)
	return cliInstallationID(id.InstallationID) && cliInstallationResource(id.Resource) && id.ClientID == "riido-cli" && err == nil && len(digest) == sha256.Size && strings.ToLower(id.KeyFingerprint) == id.KeyFingerprint
}

func (id cliInstallationIdentity) key() string {
	sum := sha256.Sum256([]byte(id.Resource + "\x00" + id.ClientID + "\x00" + id.InstallationID))
	return hex.EncodeToString(sum[:])
}

func installationPublicKey(req cliInstallationIntentRequest) (*ecdsa.PublicKey, error) {
	if req.SchemaVersion != cliInstallationIntentSchema || !req.valid() || req.PublicKeyAlgorithm != cliInstallationAlgorithm || !cliInstallationID(req.ExpectedDeviceID) || !cliInstallationID(req.ApprovedPrincipalID) {
		return nil, errCLIInstallationInvalid
	}
	key, err := base64.RawURLEncoding.DecodeString(req.PublicKey)
	if err != nil || len(key) != 65 || key[0] != 4 || base64.RawURLEncoding.EncodeToString(key) != req.PublicKey {
		return nil, errCLIInstallationInvalid
	}
	digest := sha256.Sum256(key)
	parsed, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), key)
	if err != nil || hex.EncodeToString(digest[:]) != req.KeyFingerprint {
		return nil, errCLIInstallationInvalid
	}
	return parsed, nil
}
