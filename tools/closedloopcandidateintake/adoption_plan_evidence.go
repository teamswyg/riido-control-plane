package main

type candidateAdoptionPlanEvidence struct {
	CandidateID           string         `json:"candidate_id"`
	RequiredNextArtifacts []string       `json:"required_next_artifacts"`
	AdoptionPlan          []adoptionStep `json:"adoption_plan"`
}

func adoptionPlanEvidence(item closedLoopCandidate) candidateAdoptionPlanEvidence {
	return candidateAdoptionPlanEvidence{
		CandidateID:           item.ID,
		RequiredNextArtifacts: append([]string(nil), item.RequiredNextArtifacts...),
		AdoptionPlan:          append([]adoptionStep(nil), item.AdoptionPlan...),
	}
}
