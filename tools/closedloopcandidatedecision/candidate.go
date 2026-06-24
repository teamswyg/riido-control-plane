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
	return result, nil
}

func decisionsByID(decisions []decisionRecord) map[string]decisionRecord {
	out := map[string]decisionRecord{}
	for _, decision := range decisions {
		out[decision.CandidateID] = decision
	}
	return out
}
