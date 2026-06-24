package main

type impactEvidence struct {
	Enabled                 bool                 `json:"enabled"`
	BaseRef                 string               `json:"base_ref,omitempty"`
	ChangedFileCount        int                  `json:"changed_file_count"`
	ChangedClaimCount       int                  `json:"changed_claim_count"`
	BoundSurfaceChangeCount int                  `json:"bound_surface_change_count"`
	Claims                  []impactClaim        `json:"claims,omitempty"`
	BoundSurfaces           []impactBoundSurface `json:"bound_surfaces,omitempty"`
}

type impactClaim struct {
	ID                string   `json:"id"`
	ChangedBoundFiles []string `json:"changed_bound_files"`
}

type impactBoundSurface struct {
	ID                string   `json:"id"`
	ChangedBoundFiles []string `json:"changed_bound_files"`
	ChangedEvidence   []string `json:"changed_evidence"`
}
