package main

import "encoding/json"

type rawSubject = json.RawMessage

type candidateSubject struct {
	Kind              string `json:"kind"`
	HarnessLoop       string `json:"harness_loop"`
	SourceWorkflow    string `json:"source_workflow"`
	SummaryArtifact   string `json:"summary_artifact"`
	CandidateArtifact string `json:"candidate_artifact"`
	LiveStatus        string `json:"live_status"`
	ClaimID           string `json:"claim_id"`
}
