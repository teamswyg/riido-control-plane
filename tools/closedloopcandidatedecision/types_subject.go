package main

import "encoding/json"

type rawSubject = json.RawMessage

type candidateSubjectEvidence struct {
	CandidateID string          `json:"candidate_id"`
	Kind        string          `json:"kind"`
	Subject     json.RawMessage `json:"subject"`
}
