package main

func decisionForCandidate(
	decisions map[string]decisionRecord,
	templates []decisionTemplate,
	item closedLoopCandidate,
) (decisionRecord, bool, error) {
	if decision, ok := decisions[item.ID]; ok {
		return decision, true, nil
	}
	subject, ok, err := subjectEvidence(item)
	if err != nil || !ok {
		return decisionRecord{}, false, err
	}
	for _, template := range templates {
		if template.SubjectKind == subject.Kind {
			return decisionFromTemplate(item.ID, template), true, nil
		}
	}
	return decisionRecord{}, false, nil
}

func decisionFromTemplate(candidateID string, template decisionTemplate) decisionRecord {
	return decisionRecord{
		CandidateID:  candidateID,
		Disposition:  template.Disposition,
		Priority:     template.Priority,
		Owner:        template.Owner,
		NextLoop:     template.NextLoop,
		NextArtifact: template.NextArtifact,
		ReviewBy:     template.ReviewBy,
		Reason:       template.Reason,
	}
}
