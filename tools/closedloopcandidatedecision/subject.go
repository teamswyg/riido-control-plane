package main

import (
	"encoding/json"
	"fmt"
)

func subjectEvidence(item closedLoopCandidate) (candidateSubjectEvidence, bool, error) {
	if len(item.Subject) == 0 {
		return candidateSubjectEvidence{}, false, nil
	}
	var envelope struct {
		Kind string `json:"kind"`
	}
	subject := item.Subject
	if err := json.Unmarshal(subject, &envelope); err != nil {
		return candidateSubjectEvidence{}, false, err
	}
	if envelope.Kind == "" {
		return candidateSubjectEvidence{}, false,
			fmt.Errorf("candidate %s subject must include kind", item.ID)
	}
	return candidateSubjectEvidence{CandidateID: item.ID, Kind: envelope.Kind, Subject: subject}, true, nil
}
