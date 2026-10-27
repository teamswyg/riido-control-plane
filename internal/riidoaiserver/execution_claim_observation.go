package riidoaiserver

import "context"

// Unknown is deliberately not permission to use a CLI grant. Confirmed termination,
// installation coverage and service-authenticated policy enforcement are follow-up work.
type deviceExecutionObservation string

const (
	deviceExecutionUnknown deviceExecutionObservation = "unknown"
	deviceExecutionBlocked deviceExecutionObservation = "blocked"
)

type deviceExecutionClaimReader interface {
	loadDeviceExecutionClaims(context.Context, string) ([]executionClaim, error)
}

type observeDeviceExecutionCmd struct {
	ctx      context.Context
	deviceID string
	reply    chan deviceExecutionResult
}
type deviceExecutionResult struct {
	observation deviceExecutionObservation
	err         error
}

func (s *Store) observeDeviceExecutions(ctx context.Context, deviceID string) (deviceExecutionObservation, error) {
	if reader, ok := s.operationStore.(deviceExecutionClaimReader); ok {
		claims, err := reader.loadDeviceExecutionClaims(ctx, deviceID)
		if err != nil {
			return deviceExecutionUnknown, err
		}
		return s.observeDurableExecutionClaims(ctx, claims)
	}
	reply := make(chan deviceExecutionResult, 1)
	if err := s.send(ctx, observeDeviceExecutionCmd{ctx, deviceID, reply}); err != nil {
		return deviceExecutionUnknown, err
	}
	select {
	case result := <-reply:
		return result.observation, result.err
	case <-ctx.Done():
		return deviceExecutionUnknown, ctx.Err()
	}
}

func claimMatchesNonOwnerAssignment(claim executionClaim, a Assignment) bool {
	return claim.RequesterPrincipalID != "" && claim.OwnerPrincipalID != claim.RequesterPrincipalID &&
		a.ID == claim.AssignmentID && a.AgentID == claim.AgentID && a.LeaseToken != "" &&
		leaseEpoch(a.LeaseToken) == claim.LeaseEpoch && assignmentHoldsActiveLease(a.State)
}
