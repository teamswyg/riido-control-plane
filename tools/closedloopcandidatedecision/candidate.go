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
	if err := verifyCandidateFresh(candidate, evidenceNow()); err != nil {
		return verifyResult{}, err
	}
	decisionByID := decisionsByID(m.Decisions)
	result := verifyResult{CandidateCount: candidate.CandidateCount}
	candidateIDs := []string{}
	for _, item := range candidate.Candidates {
		if err := verifyCandidateSourceRef(item, candidate); err != nil {
			return result, err
		}
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
		candidateIDs = append(candidateIDs, item.ID)
		result.CandidateSourceRefs = append(result.CandidateSourceRefs, sourceRefEvidence(item))
		subject, ok, err := subjectEvidence(item)
		if err != nil {
			return result, err
		}
		if ok {
			result.CandidateSubjects = append(result.CandidateSubjects, subject)
		}
		result.DecisionArtifacts = append(result.DecisionArtifacts, decisionArtifactEvidence{
			CandidateID:   item.ID,
			Disposition:   decision.Disposition,
			Priority:      decision.Priority,
			Owner:         decision.Owner,
			ReviewBy:      decision.ReviewBy,
			NextLoop:      decision.NextLoop,
			NextArtifact:  decision.NextArtifact,
			NextCommand:   command,
			PromotionEdge: item.PromotionEdge,
		})
	}
	if err := verifyNoOrphanDecisions(m.Decisions, candidate.Candidates, candidate.ID); err != nil {
		return result, err
	}
	result.ConsumedCandidateArtifacts = []consumedCandidateArtifact{
		consumedArtifact(path, candidate, candidateIDs),
	}
	return result, nil
}
