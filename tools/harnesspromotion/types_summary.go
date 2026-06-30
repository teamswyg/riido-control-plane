package main

type liveSummary struct {
	SchemaVersion  string      `json:"schema_version"`
	ID             string      `json:"id"`
	LiveStatus     string      `json:"live_status"`
	EvidenceClaims []liveClaim `json:"evidence_claims,omitempty"`
	Run            runRecord   `json:"run"`
	GeneratedAt    string      `json:"generated_at"`
	ExpiresAt      string      `json:"expires_at"`
}

type liveClaim struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
	Status  string `json:"status"`
}

type runRecord struct {
	ID      string `json:"id,omitempty"`
	Attempt string `json:"attempt,omitempty"`
	SHA     string `json:"sha,omitempty"`
	RefName string `json:"ref_name,omitempty"`
	Event   string `json:"event,omitempty"`
}

type verifyResult struct {
	SourceCount                     int `json:"source_count"`
	SidecarSourceCount              int `json:"sidecar_source_count"`
	LoopOwnedCandidateProducerCount int `json:"loop_owned_candidate_producer_count"`
	ClaimCount                      int `json:"claim_count"`
}
