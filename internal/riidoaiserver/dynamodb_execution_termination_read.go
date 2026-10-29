package riidoaiserver

import (
	"context"
	"encoding/json"
)

func (s *DynamoDBAssignmentOperationStore) loadExecutionTerminationReadSet(ctx context.Context, id executionTerminationIdentity) (executionTerminationReadSet, error) {
	var set executionTerminationReadSet
	item, err := s.executionTerminationGet(ctx, "DEVICE_EXECUTION#"+id.deviceID, id.key())
	if err != nil {
		return set, err
	}
	if len(item) == 0 {
		return set, errTerminationForbidden
	}
	claim, err := executionClaimFromItem(item)
	if err != nil {
		return set, err
	}
	if !id.matches(claim) {
		return set, errTerminationForbidden
	}
	set.claim = item
	set.claimValue = claim
	if claim.CoverageGeneration != "" {
		coverage, covered, err := s.loadExecutionCoverage(ctx, id.deviceID)
		if err != nil {
			return set, err
		}
		if !covered || coverage.Generation != claim.CoverageGeneration || coverage.OwnerPrincipalID != claim.OwnerPrincipalID || coverage.Generation != id.coverage.generation || coverage.CredentialGeneration != id.coverage.credential {
			return set, errTerminationForbidden
		}
		set.coverage = &coverage
	}
	set.projection, err = s.executionTerminationGet(ctx, dynamoDBAssignmentProjectionPK(id.assignmentID), dynamoDBAssignmentProjectionSK)
	if err != nil {
		return set, err
	}
	if len(set.projection) == 0 {
		return set, errTerminationForbidden
	}
	projection, err := assignmentProjectionFromDynamoDBItem(set.projection)
	if err != nil {
		return set, err
	}
	if dynamoDBStringValue(set.projection, "pk") != dynamoDBAssignmentProjectionPK(id.assignmentID) || dynamoDBStringValue(set.projection, "sk") != dynamoDBAssignmentProjectionSK || dynamoDBStringValue(set.projection, "lease_epoch") != leaseEpoch(projection.Assignment.LeaseToken) || dynamoDBStringValue(set.projection, "assignment_state") != string(projection.Assignment.State) {
		return set, errTerminationForbidden
	}
	if err := executionTerminationProjectionGuard(id, projection.Assignment); err != nil {
		return set, err
	}
	set.receipt, err = s.executionTerminationGet(ctx, "DEVICE_EXECUTION_TERMINATION#"+id.deviceID, id.key())
	if err != nil || len(set.receipt) == 0 {
		return set, err
	}
	if dynamoDBStringValue(set.receipt, "schema") != executionTerminationItemSchema || dynamoDBStringValue(set.receipt, "pk") != "DEVICE_EXECUTION_TERMINATION#"+id.deviceID || dynamoDBStringValue(set.receipt, "sk") != id.key() {
		return set, errTerminationUnavailable
	}
	if err := json.Unmarshal([]byte(dynamoDBStringValue(set.receipt, "payload")), &set.value); err != nil {
		return set, err
	}
	return set, validateExecutionTermination(set.value, claim)
}
