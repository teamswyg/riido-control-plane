package main

func pressureCandidateEvidenceFromReport(report pressureReport) pressureCandidateEvidence {
	generatedAt := report.StartedAt.Format(timeFormat)
	expiresAt := report.StartedAt.Add(hours(pressureCandidateTTLHours)).Format(timeFormat)
	candidates := pressureCandidates(report.Candidates, generatedAt, expiresAt)
	return pressureCandidateEvidence{
		SchemaVersion: pressureCandidateSchema, ID: pressureCandidateSourceID, Status: "verified",
		SourceWorkflow: pressureSourceWorkflow, LiveStatus: pressureLiveStatus,
		SourceGeneratedAt: generatedAt, SourceExpiresAt: expiresAt, Run: githubRunRecord(),
		CandidateCount: len(candidates), Candidates: candidates,
		Redaction: pressureCandidateRedact{true, true, true, true},
	}
}

func pressureCandidates(
	candidates []candidateEntry,
	generatedAt string,
	expiresAt string,
) []pressureLoopCandidate {
	out := make([]pressureLoopCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		out = append(out, pressureCandidate(candidate, generatedAt, expiresAt))
	}
	return out
}
