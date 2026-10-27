package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"strings"
)

func bumpDevicePolicyRevision(state *storeState, device string) error {
	if !cliInstallationID(device) || state.devicePolicyRevisions[device] == math.MaxUint64 {
		return errCLIInstallationUnavailable
	}
	if state.devicePolicyRevisions == nil {
		state.devicePolicyRevisions = map[string]uint64{}
	}
	state.devicePolicyRevisions[device]++
	return nil
}

func (intent cliInstallationIntent) valid() error {
	if _, err := installationPublicKey(intent.Request); err != nil {
		return err
	}
	digest, err := hex.DecodeString(intent.ChallengeDigest)
	if !cliInstallationID(intent.ID) || intent.ExpiresAt.IsZero() || err != nil || len(digest) != sha256.Size || strings.ToLower(intent.ChallengeDigest) != intent.ChallengeDigest {
		return errCLIInstallationInvalid
	}
	if payload, err := json.Marshal(intent); err != nil || len(payload) > 3<<10 {
		return errCLIInstallationInvalid
	}
	return nil
}

func (link cliInstallationLink) valid() error {
	if link.SchemaVersion != cliInstallationLinkSchema || !link.cliInstallationIdentity.valid() || !cliInstallationID(link.DeviceID) || !cliInstallationID(link.DeviceOwnerAtLink) || link.DeviceOwnerAtLink != link.LinkApprovedBy || link.LinkRevision != "1" || link.LinkedAt.IsZero() {
		return errCLIInstallationInvalid
	}
	if payload, err := json.Marshal(link); err != nil || len(payload) > 3<<10 {
		return errCLIInstallationInvalid
	}
	return nil
}
