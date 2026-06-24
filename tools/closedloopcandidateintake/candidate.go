package main

import "fmt"

func verifyCandidateFile(root string, m manifest, path string) (verifyResult, error) {
	candidate, data, err := loadCandidate(repoPath(root, path))
	if err != nil {
		return verifyResult{}, err
	}
	if err := verifyCandidateEnvelope(candidate, data); err != nil {
		return verifyResult{}, err
	}
	result := verifyResult{CandidateCount: candidate.CandidateCount}
	for _, item := range candidate.Candidates {
		if err := verifyCandidateItem(m, item); err != nil {
			return result, err
		}
		result.CandidateIDs = append(result.CandidateIDs, item.ID)
	}
	return result, nil
}

func verifyCandidateEnvelope(candidate candidateEvidence, data []byte) error {
	if candidate.SchemaVersion != candidateSchema || candidate.Status != "verified" {
		return fmt.Errorf("unexpected candidate artifact identity")
	}
	if candidate.CandidateCount != len(candidate.Candidates) {
		return fmt.Errorf("candidate_count does not match candidate list")
	}
	if !candidate.Redaction.SummaryOnly || !candidate.Redaction.NoRawSecrets {
		return fmt.Errorf("candidate artifact must be redacted")
	}
	return verifyNoRawLeak(data)
}
