package main

type manifest struct {
	SchemaVersion string     `json:"schema_version"`
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	RiidoTask     string     `json:"riido_task"`
	GeneratedDoc  string     `json:"generated_doc"`
	Workflow      string     `json:"workflow"`
	Evidence      string     `json:"evidence_artifact"`
	LinkedCD      string     `json:"linked_runtime_cd_manifest"`
	Boundaries    []boundary `json:"boundaries"`
	Rules         []string   `json:"rules"`
	Loop          loopRecord `json:"loop"`
}

type boundary struct {
	ID         string          `json:"id"`
	Owner      string          `json:"owner"`
	Scope      string          `json:"scope"`
	DoesNotOwn []string        `json:"does_not_own"`
	Evidence   []evidenceCheck `json:"evidence"`
}

type evidenceCheck struct {
	Path     string   `json:"path"`
	Contains []string `json:"contains"`
}

type loopRecord struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
