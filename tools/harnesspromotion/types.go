package main

type manifest struct {
	SchemaVersion    string            `json:"schema_version"`
	ID               string            `json:"id"`
	Title            string            `json:"title"`
	GeneratedDoc     string            `json:"generated_doc"`
	Workflow         string            `json:"workflow"`
	EvidenceArtifact string            `json:"evidence_artifact"`
	EvidenceTool     string            `json:"evidence_tool"`
	Sources          []promotionSource `json:"sources"`
	Assertions       []string          `json:"assertions"`
	Loop             evidenceLoop      `json:"loop"`
}

type promotionSource struct {
	ID                    string   `json:"id"`
	HarnessLoop           string   `json:"harness_loop"`
	SourceWorkflow        string   `json:"source_workflow"`
	SummaryArtifact       string   `json:"summary_artifact"`
	SummaryPath           string   `json:"summary_path"`
	CandidateArtifact     string   `json:"candidate_artifact"`
	CandidatePath         string   `json:"candidate_path"`
	PromotionTarget       string   `json:"promotion_target"`
	FailureStatuses       []string `json:"failure_statuses"`
	RequiredNextArtifacts []string `json:"required_next_artifacts"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
