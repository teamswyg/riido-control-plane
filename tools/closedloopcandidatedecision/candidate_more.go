package main

import "fmt"

func verifyCandidateEnvelope(candidate candidateEvidence, data []byte) error {
	if candidate.SchemaVersion != candidateSchema || candidate.Status != "verified" {
		return fmt.Errorf("unexpected candidate artifact identity")
	}
	if candidate.CandidateCount != len(candidate.Candidates) {
		return fmt.Errorf("candidate_count does not match candidate list")
	}
	if !candidate.Redaction.SummaryOnly || !candidate.Redaction.NoRawSecrets || !candidate.Redaction.NoRawEndpoints {
		return fmt.Errorf("candidate artifact must be redacted")
	}
	return verifyNoRawLeak(data)
}
