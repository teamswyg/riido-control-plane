package main

type loopCompletion struct {
	LoopID                string   `json:"loop_id"`
	Kind                  string   `json:"kind"`
	Status                string   `json:"status"`
	RequiredChecks        int      `json:"required_checks"`
	PassedChecks          int      `json:"passed_checks"`
	CompletionBasisPoints int      `json:"completion_basis_points"`
	ThresholdBasisPoints  int      `json:"threshold_basis_points"`
	MissingChecks         []string `json:"missing_checks,omitempty"`
}

type loopCompletionSummary struct {
	LoopCount                int      `json:"loop_count"`
	VerifiedLoopCount        int      `json:"verified_loop_count"`
	BelowThresholdCount      int      `json:"below_threshold_count"`
	MinCompletionBasisPoints int      `json:"min_completion_basis_points"`
	ThresholdBasisPoints     int      `json:"threshold_basis_points"`
	BelowThresholdLoopIDs    []string `json:"below_threshold_loop_ids,omitempty"`
}
