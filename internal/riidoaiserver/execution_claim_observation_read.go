package riidoaiserver

import "context"

func (s *Store) observeDurableExecutionClaims(ctx context.Context, claims []executionClaim) (deviceExecutionObservation, error) {
	projections, ok := s.operationStore.(AssignmentProjectionReader)
	leases, leaseOK := s.operationStore.(AssignmentActiveLeaseStore)
	if !ok || !leaseOK {
		return deviceExecutionUnknown, nil
	}
	for _, claim := range claims {
		projection, found, err := projections.LoadAssignmentProjection(ctx, claim.AssignmentID)
		if err != nil {
			return deviceExecutionUnknown, err
		}
		if !found || !claimMatchesNonOwnerAssignment(claim, projection.Assignment) {
			continue
		}
		lease, found, err := leases.LoadAgentActiveAssignment(ctx, claim.AgentID)
		if err != nil {
			return deviceExecutionUnknown, err
		}
		if found && lease.ActiveAssignmentID == claim.AssignmentID && lease.LeaseToken != "" && leaseEpoch(lease.LeaseToken) == claim.LeaseEpoch && !lease.Expired(s.now()) {
			return deviceExecutionBlocked, nil
		}
	}
	return deviceExecutionUnknown, nil
}

func (s *Store) observeLocalExecutionClaims(state *storeState, msg observeDeviceExecutionCmd) deviceExecutionResult {
	if err := msg.ctx.Err(); err != nil {
		return deviceExecutionResult{deviceExecutionUnknown, err}
	}
	for _, claim := range state.executionClaims {
		if claim.DeviceID != msg.deviceID {
			continue
		}
		assignment := state.assignments[claim.AssignmentID]
		if !claimMatchesNonOwnerAssignment(claim, assignment) {
			continue
		}
		expired, err := s.assignmentActiveLeaseExpired(state, assignment, s.now())
		if err != nil {
			return deviceExecutionResult{deviceExecutionUnknown, err}
		}
		if !expired {
			return deviceExecutionResult{deviceExecutionBlocked, nil}
		}
	}
	return deviceExecutionResult{deviceExecutionUnknown, nil}
}
