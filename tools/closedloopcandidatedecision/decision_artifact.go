package main

func decisionArtifactFor(
	item closedLoopCandidate,
	resolved resolvedDecision,
	command string,
) decisionArtifactEvidence {
	decision := resolved.Record
	return decisionArtifactEvidence{
		CandidateID:                 item.ID,
		Disposition:                 decision.Disposition,
		Priority:                    decision.Priority,
		Owner:                       decision.Owner,
		ReviewBy:                    decision.ReviewBy,
		NextLoop:                    decision.NextLoop,
		NextArtifact:                decision.NextArtifact,
		NextCommand:                 command,
		DecisionSource:              resolved.Source,
		DecisionTemplateSubjectKind: resolved.TemplateSubjectKind,
		PromotionEdge:               item.PromotionEdge,
	}
}
