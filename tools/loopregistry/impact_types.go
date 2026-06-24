package main

type impactEvidence struct {
	Enabled           bool          `json:"enabled"`
	BaseRef           string        `json:"base_ref,omitempty"`
	ChangedFileCount  int           `json:"changed_file_count"`
	ChangedClaimCount int           `json:"changed_claim_count"`
	Claims            []impactClaim `json:"claims,omitempty"`
}

type impactClaim struct {
	ID                string   `json:"id"`
	ChangedBoundFiles []string `json:"changed_bound_files"`
}
