package main

const (
	candidateSchema     = "riido-control-plane-closed-loop-candidates.v1"
	candidateLiveStatus = "audit_gaps"
)

func newCandidateEvidence(m manifest, deps dependencies) candidateEvidence {
	source := m.Sources[0]
	generatedAt, expiresAt := candidateWindow()
	candidates := auditCandidates(source, m, deps, generatedAt, expiresAt)
	return candidateEvidence{
		SchemaVersion:     candidateSchema,
		ID:                source.ID,
		Status:            "verified",
		SourceWorkflow:    source.SourceWorkflow,
		LiveStatus:        candidateLiveStatus,
		SourceGeneratedAt: generatedAt,
		SourceExpiresAt:   expiresAt,
		CandidateCount:    len(candidates),
		Candidates:        candidates,
		Redaction:         candidateRedaction{true, true, true, true},
	}
}

func auditCandidates(
	source candidateSource,
	m manifest,
	deps dependencies,
	generatedAt string,
	expiresAt string,
) []closedLoopCandidate {
	out := residualGapCandidates(source, m.ResidualGaps, generatedAt, expiresAt)
	out = append(out, claimCoverageCandidates(source, deps, generatedAt, expiresAt)...)
	return out
}
