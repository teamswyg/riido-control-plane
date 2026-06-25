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
	if err := verifyCandidateFresh(candidate, evidenceNow()); err != nil {
		return verifyResult{}, err
	}
	result := verifyResult{CandidateCount: candidate.CandidateCount}
	sourceIDs := []string{}
	for _, item := range candidate.Candidates {
		source, err := verifyCandidateItem(m, candidate, item)
		if err != nil {
			return result, err
		}
		sourceIDs = append(sourceIDs, source.ID)
		result.CandidateIDs = append(result.CandidateIDs, item.ID)
		result.CandidateEdges = append(result.CandidateEdges, item.PromotionEdge)
		result.CandidateSourceRefs = append(result.CandidateSourceRefs, sourceRefEvidence(item))
	}
	result.ConsumedCandidateArtifacts = []consumedCandidateArtifact{
		consumedArtifact(path, candidate, result.CandidateIDs, sourceIDs),
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

func consumedArtifact(
	path string,
	candidate candidateEvidence,
	ids []string,
	sources []string,
) consumedCandidateArtifact {
	return consumedCandidateArtifact{
		InputPath:         path,
		SourceWorkflow:    candidate.SourceWorkflow,
		LiveStatus:        candidate.LiveStatus,
		SourceGeneratedAt: candidate.SourceGeneratedAt,
		SourceExpiresAt:   candidate.SourceExpiresAt,
		CandidateCount:    candidate.CandidateCount,
		CandidateIDs:      append([]string(nil), ids...),
		SourceIDs:         uniqueStrings(sources),
	}
}
