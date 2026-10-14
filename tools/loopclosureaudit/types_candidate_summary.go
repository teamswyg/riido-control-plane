package main

type candidateSummary struct {
	CandidateArtifact string   `json:"candidate_artifact,omitempty"`
	CandidateCount    int      `json:"candidate_count"`
	CandidateIDs      []string `json:"candidate_ids"`
	CandidateSourceID string   `json:"candidate_source_id,omitempty"`
	PromotionTarget   string   `json:"candidate_promotion_target,omitempty"`
}
