package main

type candidateSourceRef struct {
	HarnessLoop       string    `json:"harness_loop"`
	SourceWorkflow    string    `json:"source_workflow"`
	SummaryArtifact   string    `json:"summary_artifact"`
	CandidateArtifact string    `json:"candidate_artifact"`
	LiveStatus        string    `json:"live_status"`
	SourceGeneratedAt string    `json:"source_generated_at"`
	SourceExpiresAt   string    `json:"source_expires_at"`
	Run               runRecord `json:"run"`
}

type candidateSourceRefEvidence struct {
	CandidateID string             `json:"candidate_id"`
	SourceRef   candidateSourceRef `json:"source_ref"`
}

type runRecord struct {
	ID      string `json:"id,omitempty"`
	Attempt string `json:"attempt,omitempty"`
	SHA     string `json:"sha,omitempty"`
	RefName string `json:"ref_name,omitempty"`
	Event   string `json:"event,omitempty"`
}
