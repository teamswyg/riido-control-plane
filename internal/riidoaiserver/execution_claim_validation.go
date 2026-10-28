package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

func (claim executionClaim) validate() error {
	for _, id := range []string{claim.DeviceID, claim.OwnerPrincipalID, claim.AssignmentID, claim.AgentID} {
		if strings.TrimSpace(id) == "" || len(id) > 512 {
			return errors.New("invalid execution claim identity")
		}
	}
	if len(claim.RequesterPrincipalID) > 512 || claim.ClaimedAt.IsZero() {
		return errors.New("invalid execution claim provenance")
	}
	if err := validateClaimCoverage(claim); err != nil {
		return err
	}
	digest, err := hex.DecodeString(claim.LeaseEpoch)
	if err != nil || len(digest) != sha256.Size || strings.ToLower(claim.LeaseEpoch) != claim.LeaseEpoch {
		return errors.New("invalid execution claim epoch")
	}
	return nil
}
