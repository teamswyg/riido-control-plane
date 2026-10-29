package riidoaiserver

import (
	"errors"
	"sort"
)

func rememberExecutionClaim(state *storeState, claim *executionClaim) error {
	if claim == nil {
		return nil
	}
	if err := claim.validate(); err != nil {
		return err
	}
	if old, ok := state.executionClaims[claim.key()]; ok && old != *claim {
		return errors.New("execution claim cannot be replaced")
	}
	if state.executionClaims == nil {
		state.executionClaims = map[string]executionClaim{}
	}
	if _, exists := state.executionClaims[claim.key()]; !exists {
		if err := bumpDevicePolicyRevision(state, claim.DeviceID); err != nil {
			return err
		}
	}
	state.executionClaims[claim.key()] = *claim
	return nil
}

func executionClaimsFromState(state *storeState) []executionClaim {
	claims := make([]executionClaim, 0, len(state.executionClaims))
	for _, claim := range state.executionClaims {
		claims = append(claims, claim)
	}
	sort.Slice(claims, func(i, j int) bool { return claims[i].key() < claims[j].key() })
	return claims
}

func executionClaimForOperation(state *storeState, kind AssignmentOperationType, assignment Assignment) *executionClaim {
	if kind != AssignmentOperationPollStart {
		return nil
	}
	claim, ok := state.executionClaims[assignment.ID+"#"+leaseEpoch(assignment.LeaseToken)]
	if !ok {
		return nil
	}
	return &claim
}

func validateOperationExecutionClaim(record AssignmentOperationRecord) error {
	claim := record.ExecutionClaim
	if claim == nil {
		return nil
	}
	if err := claim.validate(); err != nil {
		return err
	}
	a := record.Assignment
	if claim.CoverageGeneration != "" && (claim.RuntimeProvider != a.RuntimeProvider || claim.TerminationScope != coverageTerminationScope(a)) {
		return errors.New("execution coverage runtime does not match leased operation")
	}
	if record.OperationType != AssignmentOperationPollStart || a.State != AssignmentLeased || a.LeaseToken == "" ||
		claim.AssignmentID != a.ID || claim.AgentID != a.AgentID || claim.RequesterPrincipalID != a.CreatedBy ||
		claim.LeaseEpoch != leaseEpoch(a.LeaseToken) || !claim.ClaimedAt.Equal(a.UpdatedAt) {
		return errors.New("execution claim does not match leased operation")
	}
	return nil
}
