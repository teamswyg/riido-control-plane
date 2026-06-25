package main

type manifest struct {
	SchemaVersion    string         `json:"schema_version"`
	ID               string         `json:"id"`
	Title            string         `json:"title"`
	GeneratedDoc     string         `json:"generated_doc"`
	Workflow         string         `json:"workflow"`
	EvidenceArtifact string         `json:"evidence_artifact"`
	EvidenceTool     string         `json:"evidence_tool"`
	Sources          []intakeSource `json:"sources"`
	Assertions       []string       `json:"assertions"`
	Loop             evidenceLoop   `json:"loop"`
}

type intakeSource struct {
	ID                    string   `json:"id"`
	CandidateArtifact     string   `json:"candidate_artifact"`
	SourceWorkflow        string   `json:"source_workflow"`
	ProducerManifest      string   `json:"producer_manifest"`
	LoopRegistryManifest  string   `json:"loop_registry_manifest"`
	EvidenceGraphManifest string   `json:"evidence_graph_manifest"`
	HarnessLoop           string   `json:"harness_loop"`
	PromotionTarget       string   `json:"promotion_target"`
	RequiredNextArtifacts []string `json:"required_next_artifacts"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
