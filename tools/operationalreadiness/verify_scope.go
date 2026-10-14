package main

import "fmt"

func verifyCompletionScope(check readinessCheck) error {
	switch check.CompletionScope {
	case completionScopeInternal, completionScopeExternal:
	default:
		return fmt.Errorf("readiness check %s has invalid completion_scope %q",
			check.ID, check.CompletionScope)
	}
	if check.ScopeReason == "" {
		return fmt.Errorf("readiness check %s must explain completion scope", check.ID)
	}
	return nil
}

func verifyInternalCompletion(checks []readinessCheck) error {
	evidence := newCompletionEvidence(checks)
	if evidence.InternalCheckCount == 0 {
		return fmt.Errorf("readiness must have at least one internal check")
	}
	if evidence.InternalCompletenessBasisPoints < evidence.ThresholdBasisPoints {
		return fmt.Errorf("internal readiness completeness %d below threshold %d",
			evidence.InternalCompletenessBasisPoints, evidence.ThresholdBasisPoints)
	}
	return nil
}
