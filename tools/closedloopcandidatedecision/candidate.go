package main

import "fmt"

func verifyCandidateDecisions(root string, m manifest, path string) (verifyResult, error) {
	candidate, data, err := loadCandidate(repoPath(root, path))
	if err != nil {
		return verifyResult{}, err
	}
	if err := verifyCandidateEnvelope(candidate, data); err != nil {
		return verifyResult{}, err
	}
	decisionByID := decisionsByID(m.Decisions)
	result := verifyResult{CandidateCount: candidate.CandidateCount}
	for _, item := range candidate.Candidates {
		if _, ok := decisionByID[item.ID]; !ok {
			return result, fmt.Errorf("candidate %s has no decision record", item.ID)
		}
		result.DecisionIDs = append(result.DecisionIDs, item.ID)
	}
	if err := verifyNoOrphanDecisions(m.Decisions, candidate.Candidates); err != nil {
		return result, err
	}
	return result, nil
}

func decisionsByID(decisions []decisionRecord) map[string]decisionRecord {
	out := map[string]decisionRecord{}
	for _, decision := range decisions {
		out[decision.CandidateID] = decision
	}
	return out
}

func verifyNoOrphanDecisions(decisions []decisionRecord, candidates []closedLoopCandidate) error {
	candidateByID := map[string]bool{}
	for _, item := range candidates {
		candidateByID[item.ID] = true
	}
	for _, decision := range decisions {
		if !candidateByID[decision.CandidateID] {
			return fmt.Errorf("decision %s has no matching candidate", decision.CandidateID)
		}
	}
	return nil
}
