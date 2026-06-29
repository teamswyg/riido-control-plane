package main

import "time"

func newCandidateEvidence(m manifest, e evidence, now time.Time) candidateEvidence {
	source := readinessCandidateSource(m)
	generatedAt := now.UTC().Format(time.RFC3339)
	expiresAt := now.UTC().Add(24 * time.Hour).Format(time.RFC3339)
	run := githubRunRecord()
	candidates := stalePartialCandidates(source, e, generatedAt, expiresAt, run)
	return candidateEvidence{
		SchemaVersion:     candidateSchema,
		ID:                source.ID,
		Status:            "verified",
		SourceWorkflow:    source.SourceWorkflow,
		LiveStatus:        candidateLiveStatus(candidates),
		SourceGeneratedAt: generatedAt,
		SourceExpiresAt:   expiresAt,
		Run:               run,
		CandidateCount:    len(candidates),
		Candidates:        candidates,
		Redaction:         candidateRedaction{true, true, true, true},
	}
}

func stalePartialCandidates(
	source producerSource,
	e evidence,
	generatedAt string,
	expiresAt string,
	run runRecord,
) []closedLoopCandidate {
	candidates := []closedLoopCandidate{}
	for _, partial := range e.PartialChecks {
		if partial.Stale {
			candidates = append(candidates, newStalePartialCandidate(
				source, partial, generatedAt, expiresAt, run))
		}
	}
	return candidates
}

func candidateLiveStatus(candidates []closedLoopCandidate) string {
	if len(candidates) == 0 {
		return "verified"
	}
	return "stale_partial"
}
