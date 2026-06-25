package main

type manifest struct {
	SchemaVersion         string    `json:"schema_version"`
	ID                    string    `json:"id"`
	Title                 string    `json:"title"`
	GeneratedDoc          string    `json:"generated_doc"`
	Workflow              string    `json:"workflow"`
	EvidenceArtifact      string    `json:"evidence_artifact"`
	EvidenceTool          string    `json:"evidence_tool"`
	BenchmarkCommand      string    `json:"benchmark_command"`
	LocalPressureCommand  string    `json:"local_pressure_command"`
	ManualPressureCommand string    `json:"manual_pressure_command"`
	RaceCommand           string    `json:"race_command"`
	PprofCommands         []string  `json:"pprof_commands"`
	RequiredCategories    []string  `json:"required_categories"`
	Surfaces              []surface `json:"surfaces"`
	Assertions            []string  `json:"assertions"`
	Loop                  loopSpec  `json:"loop"`
}

type surface struct {
	ID        string   `json:"id"`
	Category  string   `json:"category"`
	Risk      string   `json:"risk"`
	Files     []string `json:"files"`
	Patterns  []string `json:"patterns"`
	Candidate string   `json:"candidate"`
}

type loopSpec struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
