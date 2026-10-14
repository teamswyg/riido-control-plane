package main

type consumedCandidateArtifact struct {
	InputPath         string   `json:"input_path"`
	SourceWorkflow    string   `json:"source_workflow"`
	LiveStatus        string   `json:"live_status"`
	SourceGeneratedAt string   `json:"source_generated_at"`
	SourceExpiresAt   string   `json:"source_expires_at"`
	CandidateCount    int      `json:"candidate_count"`
	CandidateIDs      []string `json:"candidate_ids"`
	SourceIDs         []string `json:"source_ids"`
}
