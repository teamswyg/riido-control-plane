package main

type promotionResult struct {
	ID                string   `json:"id"`
	CandidateArtifact string   `json:"candidate_artifact"`
	LiveStatus        string   `json:"live_status"`
	CandidateCount    int      `json:"candidate_count"`
	CandidateIDs      []string `json:"candidate_ids"`
	SourceGeneratedAt string   `json:"source_generated_at"`
	SourceExpiresAt   string   `json:"source_expires_at"`
}

func newPromotionResult(path string, evidence candidateEvidence) *promotionResult {
	return &promotionResult{
		ID:                evidence.ID,
		CandidateArtifact: path,
		LiveStatus:        evidence.LiveStatus,
		CandidateCount:    evidence.CandidateCount,
		CandidateIDs:      candidateIDs(evidence.Candidates),
		SourceGeneratedAt: evidence.SourceGeneratedAt,
		SourceExpiresAt:   evidence.SourceExpiresAt,
	}
}

func candidateIDs(candidates []closedLoopCandidate) []string {
	out := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		out = append(out, candidate.ID)
	}
	return out
}
