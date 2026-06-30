package main

type partialPromotion struct {
	CandidateArtifact string   `json:"candidate_artifact"`
	CandidateCount    int      `json:"candidate_count"`
	CandidateIDs      []string `json:"candidate_ids"`
	StalePartialCount int      `json:"stale_partial_count"`
	StalePartialIDs   []string `json:"stale_partial_ids"`
	StaleAfterDays    int      `json:"stale_after_days"`
}
