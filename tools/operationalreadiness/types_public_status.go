package main

type publicStatus struct {
	Overall              string `json:"overall"`
	Visibility           string `json:"visibility"`
	StatusPage           string `json:"status_page"`
	RawLogsIncluded      bool   `json:"raw_logs_included"`
	SecretsIncluded      bool   `json:"secrets_included"`
	EndpointDetails      string `json:"endpoint_details"`
	P0CycleCount         int    `json:"p0_cycle_count"`
	P0PartialCount       int    `json:"p0_partial_count"`
	PartialCount         int    `json:"partial_count"`
	StalePartialCount    int    `json:"stale_partial_count"`
	ClosedLoopCandidates int    `json:"closed_loop_candidates"`
	NextArtifact         string `json:"next_artifact"`
}
