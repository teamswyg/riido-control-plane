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
		if err := verifyAdoptionPlan(item); err != nil {
			return result, err
		}
		if err := verifyCandidatePromotionEdge(item); err != nil {
			return result, err
		}
		decision, ok := decisionByID[item.ID]
		if !ok {
			return result, fmt.Errorf("candidate %s has no decision record", item.ID)
		}
		if err := verifyDecisionNextArtifact(item, decision); err != nil {
			return result, err
		}
		command, ok := adoptionCommandFor(item, decision.NextArtifact)
		if !ok {
			return result, fmt.Errorf("candidate %s has no command for next_artifact %s", item.ID, decision.NextArtifact)
		}
		result.DecisionIDs = append(result.DecisionIDs, item.ID)
		result.DecisionArtifacts = append(result.DecisionArtifacts, decisionArtifactEvidence{
			CandidateID:   item.ID,
			NextArtifact:  decision.NextArtifact,
			NextCommand:   command,
			PromotionEdge: item.PromotionEdge,
		})
	}
	if err := verifyNoOrphanDecisions(m.Decisions, candidate.Candidates); err != nil {
		return result, err
	}
	return result, nil
}

func verifyDecisionNextArtifact(candidate closedLoopCandidate, decision decisionRecord) error {
	if !containsString(candidate.RequiredNextArtifacts, decision.NextArtifact) {
		return fmt.Errorf("candidate %s decision next_artifact %s is not required by candidate artifact", candidate.ID, decision.NextArtifact)
	}
	return nil
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
