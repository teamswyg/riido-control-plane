package main

type manifest struct {
	SchemaVersion  string       `json:"schema_version"`
	ID             string       `json:"id"`
	Title          string       `json:"title"`
	RiidoTask      string       `json:"riido_task"`
	GeneratedDoc   string       `json:"generated_doc"`
	Workflow       string       `json:"workflow"`
	Evidence       string       `json:"evidence_artifact"`
	SourceDir      string       `json:"source_dir"`
	Entries        []entry      `json:"entries"`
	Rules          []string     `json:"rules"`
	Loop           loopEvidence `json:"loop"`
	NonConfigFacts []string     `json:"non_config_facts"`
}

type entry struct {
	Name        string `json:"name"`
	Default     string `json:"default"`
	Owner       string `json:"owner"`
	Sensitivity string `json:"sensitivity"`
	Meaning     string `json:"meaning"`
}

type loopEvidence struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
