package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	CoverageGeneration   string    `json:"coverage_generation,omitempty"`
	CredentialGeneration string    `json:"credential_generation,omitempty"`
	RuntimeProvider      string    `json:"runtime_provider,omitempty"`
	TerminationScope     string    `json:"termination_scope,omitempty"`
}

type (
	authenticatedClaimDevice struct {
		deviceID, ownerPrincipalID string
		coverage                   executionCoverageAuth
	}
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
	claim := &executionClaim{
		DeviceID: device.deviceID, OwnerPrincipalID: device.ownerPrincipalID,
		RequesterPrincipalID: assignment.CreatedBy, AssignmentID: assignment.ID, AgentID: assignment.AgentID,
		LeaseEpoch: leaseEpoch(assignment.LeaseToken), ClaimedAt: assignment.UpdatedAt,
	}
	if admission := executionAdmissionFromContext(ctx); admission != nil && admission.coverage != nil {
		claim.CoverageGeneration, claim.CredentialGeneration = admission.coverage.Generation, admission.coverage.CredentialGeneration
		claim.RuntimeProvider, claim.TerminationScope = assignment.RuntimeProvider, coverageTerminationScope(assignment)
	}
	return claim
}

func (claim executionClaim) key() string { return claim.AssignmentID + "#" + claim.LeaseEpoch }
