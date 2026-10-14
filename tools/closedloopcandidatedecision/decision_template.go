package main

func decisionForCandidate(
	decisions map[string]decisionRecord,
	templates []decisionTemplate,
	item closedLoopCandidate,
) (resolvedDecision, bool, error) {
	if decision, ok := decisions[item.ID]; ok {
		return resolvedDecision{Record: decision, Source: decisionSourceRecord}, true, nil
	}
	subject, ok, err := subjectEvidence(item)
	if err != nil || !ok {
		return resolvedDecision{}, false, err
	}
	for _, template := range templates {
		if template.SubjectKind == subject.Kind {
			return resolvedDecision{
				Record:              decisionFromTemplate(item.ID, template),
				Source:              decisionSourceTemplate,
				TemplateSubjectKind: template.SubjectKind,
			}, true, nil
		}
	}
	return resolvedDecision{}, false, nil
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
