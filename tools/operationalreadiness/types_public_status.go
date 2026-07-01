package main

type publicStatus struct {
	Overall              string                 `json:"overall"`
	Visibility           string                 `json:"visibility"`
	StatusPage           string                 `json:"status_page"`
	GeneratedAt          string                 `json:"generated_at"`
	ExpiresAt            string                 `json:"expires_at"`
	EvidenceTTLHours     int                    `json:"evidence_ttl_hours"`
	SourceWorkflow       string                 `json:"source_workflow"`
	SourceCommit         string                 `json:"source_commit"`
	SourceRunID          string                 `json:"source_run_id"`
	SourceRunURL         string                 `json:"source_run_url"`
	RawLogsIncluded      bool                   `json:"raw_logs_included"`
	SecretsIncluded      bool                   `json:"secrets_included"`
	EndpointDetails      string                 `json:"endpoint_details"`
	P0CycleCount         int                    `json:"p0_cycle_count"`
	P0PartialCount       int                    `json:"p0_partial_count"`
	PartialCount         int                    `json:"partial_count"`
	StalePartialCount    int                    `json:"stale_partial_count"`
	ClosedLoopCandidates int                    `json:"closed_loop_candidates"`
	BlockingCategories   []publicStatusCategory `json:"blocking_categories"`
	NextArtifact         string                 `json:"next_artifact"`
}

type publicStatusCategory struct {
	Category          string `json:"category"`
	PartialCount      int    `json:"partial_count"`
	StalePartialCount int    `json:"stale_partial_count"`
}
