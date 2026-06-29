package main

type manifest struct {
	SchemaVersion      string           `json:"schema_version"`
	ID                 string           `json:"id"`
	Title              string           `json:"title"`
	GeneratedDoc       string           `json:"generated_doc"`
	Workflow           string           `json:"workflow"`
	EvidenceArtifact   string           `json:"evidence_artifact"`
	EvidenceTool       string           `json:"evidence_tool"`
	RequiredCategories []string         `json:"required_categories"`
	Checks             []readinessCheck `json:"checks"`
	Loop               loopSpec         `json:"loop"`
}

type readinessCheck struct {
	ID           string        `json:"id"`
	Date         string        `json:"date"`
	Category     string        `json:"category"`
	Status       string        `json:"status"`
	Title        string        `json:"title"`
	Measurements []measurement `json:"measurements"`
	EvidenceRefs []evidenceRef `json:"evidence_refs"`
	NextArtifact string        `json:"next_artifact"`
	NextCommand  string        `json:"next_command"`
}

type measurement struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Signal      string `json:"signal"`
	EvidenceRef string `json:"evidence_ref,omitempty"`
}

type evidenceRef struct {
	Kind string `json:"kind"`
	Path string `json:"path"`
}

type loopSpec struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
