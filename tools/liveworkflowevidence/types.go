package main

type manifest struct {
	SchemaVersion string         `json:"schema_version"`
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	GeneratedDoc  string         `json:"generated_doc"`
	Workflow      string         `json:"workflow"`
	Evidence      string         `json:"evidence_artifact"`
	Workflows     []workflowSpec `json:"workflows"`
	Assertions    []string       `json:"assertions"`
	Loop          loopRecord     `json:"loop"`
}

type workflowSpec struct {
	ID              string   `json:"id"`
	Path            string   `json:"path"`
	SummaryArtifact string   `json:"summary_artifact"`
	SummaryPath     string   `json:"summary_path"`
	SensitiveInputs []string `json:"sensitive_inputs"`
	AllowedFields   []string `json:"allowed_summary_fields"`
}

type loopRecord struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
