package main

type manifest struct {
	SchemaVersion string     `json:"schema_version"`
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	RiidoTask     string     `json:"riido_task"`
	GeneratedDoc  string     `json:"generated_doc"`
	Workflow      string     `json:"workflow"`
	Evidence      string     `json:"evidence_artifact"`
	Questions     []question `json:"questions"`
	Loop          loopRecord `json:"loop"`
}

type question struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Area         string `json:"area"`
	Owner        string `json:"owner"`
	Question     string `json:"question"`
	Stance       string `json:"stance"`
	Resolution   string `json:"resolution,omitempty"`
	NextArtifact string `json:"next_artifact"`
	NextCommand  string `json:"next_command,omitempty"`
}

type loopRecord struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
