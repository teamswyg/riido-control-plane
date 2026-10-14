package main

type impactEvidence struct {
	Enabled                 bool                 `json:"enabled"`
	BaseRef                 string               `json:"base_ref,omitempty"`
	ChangedFileCount        int                  `json:"changed_file_count"`
	ChangedFiles            []string             `json:"changed_files,omitempty"`
	AddedClaimCount         int                  `json:"added_claim_count"`
	ChangedClaimCount       int                  `json:"changed_claim_count"`
	RemovedClaimCount       int                  `json:"removed_claim_count"`
	BoundSurfaceChangeCount int                  `json:"bound_surface_change_count"`
	TargetVerifierPlan      *targetVerifierPlan  `json:"target_verifier_plan,omitempty"`
	AddedClaims             []impactClaim        `json:"added_claims,omitempty"`
	Claims                  []impactClaim        `json:"claims,omitempty"`
	RemovedClaims           []impactClaim        `json:"removed_claims,omitempty"`
	BoundSurfaces           []impactBoundSurface `json:"bound_surfaces,omitempty"`
	Violations              []impactViolation    `json:"violations,omitempty"`
}

type impactClaim struct {
	ID                       string   `json:"id"`
	ChangedBoundFiles        []string `json:"changed_bound_files"`
	ChangedEvidence          []string `json:"changed_evidence"`
	ChangedReasoningEvidence []string `json:"changed_reasoning_evidence"`
}

type impactBoundSurface struct {
	ID                       string   `json:"id"`
	ChangedBoundFiles        []string `json:"changed_bound_files"`
	ChangedEvidence          []string `json:"changed_evidence"`
	ChangedReasoningEvidence []string `json:"changed_reasoning_evidence"`
}

type impactViolation struct {
	ClaimID                   string   `json:"claim_id"`
	Scope                     string   `json:"scope"`
	Reason                    string   `json:"reason"`
	RequiredBoundFiles        []string `json:"required_bound_files"`
	RequiredEvidence          []string `json:"required_evidence"`
	RequiredReasoningEvidence []string `json:"required_reasoning_evidence"`
	ChangedBoundFiles         []string `json:"changed_bound_files,omitempty"`
	ChangedEvidence           []string `json:"changed_evidence,omitempty"`
	ChangedReasoningEvidence  []string `json:"changed_reasoning_evidence,omitempty"`
}
