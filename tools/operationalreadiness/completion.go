package main

const (
	completionScopeInternal        = "internal"
	completionScopeExternal        = "external"
	internalCompletionThresholdBps = 9000
)

type completionEvidence struct {
	InternalCheckCount              int      `json:"internal_check_count"`
	InternalCoveredCount            int      `json:"internal_covered_count"`
	InternalPartialCount            int      `json:"internal_partial_count"`
	ExternalExcludedCount           int      `json:"external_excluded_count"`
	ExternalPartialCount            int      `json:"external_partial_count"`
	InternalCompletenessBasisPoints int      `json:"internal_completeness_basis_points"`
	ThresholdBasisPoints            int      `json:"threshold_basis_points"`
	Status                          string   `json:"status"`
	ExternalExcludedChecks          []string `json:"external_excluded_checks"`
}

func newCompletionEvidence(checks []readinessCheck) completionEvidence {
	var e completionEvidence
	e.ThresholdBasisPoints = internalCompletionThresholdBps
	e.ExternalExcludedChecks = []string{}
	for _, check := range checks {
		if check.CompletionScope == completionScopeExternal {
			e.ExternalExcludedCount++
			e.ExternalExcludedChecks = append(e.ExternalExcludedChecks, check.ID)
			if check.Status == "partial" {
				e.ExternalPartialCount++
			}
			continue
		}
		e.InternalCheckCount++
		if check.Status == "covered" {
			e.InternalCoveredCount++
		}
		if check.Status == "partial" {
			e.InternalPartialCount++
		}
	}
	if e.InternalCheckCount > 0 {
		e.InternalCompletenessBasisPoints = e.InternalCoveredCount * 10000 / e.InternalCheckCount
	}
	e.Status = "below_threshold"
	if e.InternalCompletenessBasisPoints >= e.ThresholdBasisPoints {
		e.Status = "verified"
	}
	return e
}
