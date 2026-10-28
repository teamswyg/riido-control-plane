package riidoaiserver

import "context"

func (s *DynamoDBAssignmentOperationStore) installationExecutionObservation(ctx context.Context, device string) (deviceExecutionObservation, error) {
	claims, err := s.loadDeviceExecutionClaims(ctx, device)
	if err != nil {
		return deviceExecutionUnknown, err
	}
	for _, claim := range claims {
		projection, found, err := s.LoadAssignmentProjection(ctx, claim.AssignmentID)
		if err != nil {
			return deviceExecutionUnknown, err
		}
		if !found || !claimMatchesNonOwnerAssignment(claim, projection.Assignment) {
			continue
		}
		lease, found, err := s.LoadAgentActiveAssignment(ctx, claim.AgentID)
		if err != nil {
			return deviceExecutionUnknown, err
		}
		if found && lease.ActiveAssignmentID == claim.AssignmentID && lease.LeaseToken != "" &&
			leaseEpoch(lease.LeaseToken) == claim.LeaseEpoch && !lease.Expired(s.now()) {
			return deviceExecutionBlocked, nil
		}
	}
	return deviceExecutionUnknown, nil
}
