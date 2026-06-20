package main

type manifest struct {
	SchemaVersion    string       `json:"schema_version"`
	ID               string       `json:"id"`
	Title            string       `json:"title"`
	RiidoTasks       []string     `json:"riido_tasks"`
	GeneratedDoc     string       `json:"generated_doc"`
	Workflow         string       `json:"workflow"`
	EvidenceArtifact string       `json:"evidence_artifact"`
	OwnerPackage     string       `json:"owner_package"`
	SharedContracts  []string     `json:"shared_contracts"`
	FocusedWorkflows []string     `json:"focused_workflows"`
	Boundaries       []boundary   `json:"boundaries"`
	RequiredPhrases  []phrase     `json:"required_phrases"`
	Loop             evidenceLoop `json:"loop"`
	NonGoals         []string     `json:"non_goals"`
}

type boundary struct {
	ID           string        `json:"id"`
	Summary      string        `json:"summary"`
	Workflow     string        `json:"workflow"`
	SourceChecks []sourceCheck `json:"source_checks"`
}

type sourceCheck struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Contains []string `json:"contains"`
}

type phrase struct {
	File     string `json:"file"`
	Contains string `json:"contains"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
