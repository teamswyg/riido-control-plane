package main

func subjectKindsByCandidate(
	subjects []candidateSubjectEvidence,
) map[string]string {
	out := map[string]string{}
	for _, subject := range subjects {
		if subject.CandidateID != "" && subject.Kind != "" {
			out[subject.CandidateID] = subject.Kind
		}
	}
	return out
}
