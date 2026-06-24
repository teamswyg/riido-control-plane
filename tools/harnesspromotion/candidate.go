package main

func buildCandidateEvidence(source promotionSource, summary liveSummary) candidateEvidence {
	candidates := candidatesForSummary(source, summary)
	return candidateEvidence{
		SchemaVersion:  candidateSchema,
		ID:             source.ID,
		Status:         "verified",
		SourceWorkflow: source.SourceWorkflow,
		LiveStatus:     summary.LiveStatus,
		Run:            summary.Run,
		CandidateCount: len(candidates),
		Candidates:     candidates,
		Redaction:      candidateRedaction{true, true, true, true},
	}
}

func candidatesForSummary(source promotionSource, summary liveSummary) []closedLoopCandidate {
	if !isFailureStatus(source, summary.LiveStatus) {
		return nil
	}
	if len(summary.EvidenceClaims) == 0 {
		return []closedLoopCandidate{newCandidate(source, "workflow", "workflow failed")}
	}
	candidates := []closedLoopCandidate{}
	for _, claim := range summary.EvidenceClaims {
		if claim.Status != "verified" {
			candidates = append(candidates, newCandidate(source, claim.ID, claim.Summary))
		}
	}
	return candidates
}

func isFailureStatus(source promotionSource, status string) bool {
	for _, candidate := range source.FailureStatuses {
		if candidate == status {
			return true
		}
	}
	return false
}
