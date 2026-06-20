package main

type manifest struct {
	SchemaVersion string     `json:"schema_version"`
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	RiidoTask     string     `json:"riido_task"`
	GeneratedDoc  string     `json:"generated_doc"`
	Workflow      string     `json:"workflow"`
	Evidence      string     `json:"evidence_artifact"`
	Gates         []gate     `json:"gates"`
	Loop          loopRecord `json:"loop"`
	NonGoals      []string   `json:"non_goals"`
}

type gate struct {
	ID       string   `json:"id"`
	Summary  string   `json:"summary"`
	Contains []string `json:"contains"`
}

type loopRecord struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
