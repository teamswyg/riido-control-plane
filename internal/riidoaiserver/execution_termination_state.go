package riidoaiserver

import (
	"errors"
	"sort"
)

func executionTerminationProjectionGuard(id executionTerminationIdentity, assignment Assignment) error {
	if assignment.ID != id.assignmentID || assignment.LeaseToken == "" {
		return errTerminationForbidden
	}
	if leaseEpoch(assignment.LeaseToken) == id.epoch && !isTerminal(assignment.State) {
		return errTerminationConflict
	}
	return nil
}

func validateExecutionTermination(receipt executionTerminationReceipt, claim executionClaim) error {
	if receipt.SchemaVersion != executionTerminationSchema || receipt.Kind != executionTerminationKind || receipt.AssignmentID != claim.AssignmentID || receipt.LeaseEpoch != claim.LeaseEpoch || receipt.RecordedAt.IsZero() {
		return errors.New("invalid execution termination receipt")
	}
	return nil
}

func rememberExecutionTermination(state *storeState, receipt executionTerminationReceipt) error {
	key := receipt.AssignmentID + "#" + receipt.LeaseEpoch
	claim, ok := state.executionClaims[key]
	if !ok {
		return errors.New("execution termination claim missing")
	}
	if err := validateExecutionTermination(receipt, claim); err != nil {
		return err
	}
	if old, exists := state.executionTerminations[key]; exists && old != receipt {
		return errTerminationConflict
	}
	state.executionTerminations[key] = receipt
	return nil
}

func executionTerminationsFromState(state *storeState) []executionTerminationReceipt {
	receipts := make([]executionTerminationReceipt, 0, len(state.executionTerminations))
	for _, receipt := range state.executionTerminations {
		receipts = append(receipts, receipt)
	}
	sort.Slice(receipts, func(i, j int) bool {
		return receipts[i].AssignmentID+"#"+receipts[i].LeaseEpoch < receipts[j].AssignmentID+"#"+receipts[j].LeaseEpoch
	})
	return receipts
}
