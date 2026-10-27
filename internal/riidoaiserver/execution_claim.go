package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

// Immutable server-authenticated claim history. This is not an authorization verdict.
type executionClaim struct {
	DeviceID             string    `json:"device_id"`
	OwnerPrincipalID     string    `json:"owner_principal_id"`
	RequesterPrincipalID string    `json:"requester_principal_id"`
	AssignmentID         string    `json:"assignment_id"`
	AgentID              string    `json:"agent_id"`
	LeaseEpoch           string    `json:"lease_epoch"`
	ClaimedAt            time.Time `json:"claimed_at"`
}

type (
	authenticatedClaimDevice    struct{ deviceID, ownerPrincipalID string }
	authenticatedClaimDeviceKey struct{}
)

func claimDeviceFromContext(ctx context.Context) (authenticatedClaimDevice, bool) {
	device, ok := ctx.Value(authenticatedClaimDeviceKey{}).(authenticatedClaimDevice)
	return device, ok && device.deviceID != "" && device.ownerPrincipalID != ""
}

func leaseEpoch(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

func newExecutionClaim(ctx context.Context, assignment Assignment) *executionClaim {
	device, ok := claimDeviceFromContext(ctx)
	if !ok {
		return nil
	}
	return &executionClaim{
		DeviceID: device.deviceID, OwnerPrincipalID: device.ownerPrincipalID,
		RequesterPrincipalID: assignment.CreatedBy, AssignmentID: assignment.ID, AgentID: assignment.AgentID,
		LeaseEpoch: leaseEpoch(assignment.LeaseToken), ClaimedAt: assignment.UpdatedAt,
	}
}

func (claim executionClaim) key() string { return claim.AssignmentID + "#" + claim.LeaseEpoch }

func (claim executionClaim) validate() error {
	for _, id := range []string{claim.DeviceID, claim.OwnerPrincipalID, claim.AssignmentID, claim.AgentID} {
		if strings.TrimSpace(id) == "" || len(id) > 512 {
			return errors.New("invalid execution claim identity")
		}
	}
	if len(claim.RequesterPrincipalID) > 512 || claim.ClaimedAt.IsZero() {
		return errors.New("invalid execution claim provenance")
	}
	digest, err := hex.DecodeString(claim.LeaseEpoch)
	if err != nil || len(digest) != sha256.Size || strings.ToLower(claim.LeaseEpoch) != claim.LeaseEpoch {
		return errors.New("invalid execution claim epoch")
	}
	return nil
}
