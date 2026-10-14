package main

type manifest struct {
	SchemaVersion    string       `json:"schema_version"`
	ID               string       `json:"id"`
	Title            string       `json:"title"`
	RiidoTask        string       `json:"riido_task"`
	RiidoTaskTitle   string       `json:"riido_task_title,omitempty"`
	GeneratedDoc     string       `json:"generated_doc"`
	Workflow         string       `json:"workflow"`
	EvidenceArtifact string       `json:"evidence_artifact"`
	Intro            []string     `json:"intro"`
	Sections         []section    `json:"sections"`
	Assertions       []string     `json:"assertions"`
	Loop             evidenceLoop `json:"loop"`
}

type section struct {
	Level int      `json:"level"`
	Title string   `json:"title"`
	Kind  string   `json:"kind"`
	Body  []string `json:"body"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
