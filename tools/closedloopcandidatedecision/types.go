package main

type manifest struct {
	SchemaVersion        string             `json:"schema_version"`
	ID                   string             `json:"id"`
	Title                string             `json:"title"`
	GeneratedDoc         string             `json:"generated_doc"`
	Workflow             string             `json:"workflow"`
	EvidenceArtifact     string             `json:"evidence_artifact"`
	CommandArtifact      string             `json:"command_artifact"`
	EvidenceTool         string             `json:"evidence_tool"`
	IntakeManifest       string             `json:"intake_manifest"`
	LoopRegistryManifest string             `json:"loop_registry_manifest"`
	Decisions            []decisionRecord   `json:"decisions"`
	DecisionTemplates    []decisionTemplate `json:"decision_templates,omitempty"`
	Assertions           []string           `json:"assertions"`
	Loop                 evidenceLoop       `json:"loop"`
}

type decisionRecord struct {
	CandidateID  string `json:"candidate_id"`
	Disposition  string `json:"disposition"`
	Priority     string `json:"priority"`
	Owner        string `json:"owner"`
	NextLoop     string `json:"next_loop"`
	NextArtifact string `json:"next_artifact"`
	ReviewBy     string `json:"review_by,omitempty"`
	Reason       string `json:"reason"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
