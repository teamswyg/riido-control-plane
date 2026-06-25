package main

type manifest struct {
	SchemaVersion     string    `json:"schema_version"`
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	GeneratedDoc      string    `json:"generated_doc"`
	Workflow          string    `json:"workflow"`
	EvidenceArtifact  string    `json:"evidence_artifact"`
	BenchmarkArtifact string    `json:"benchmark_artifact"`
	SummaryArtifact   string    `json:"summary_artifact"`
	CandidateArtifact string    `json:"candidate_artifact"`
	EvidenceTool      string    `json:"evidence_tool"`
	BenchmarkCommand  string    `json:"benchmark_command"`
	RaceCommand       string    `json:"race_command"`
	PprofCommand      string    `json:"pprof_command"`
	LiveLoadCommand   string    `json:"live_load_command"`
	HotPaths          []hotPath `json:"hot_paths"`
	Assertions        []string  `json:"assertions"`
	Loop              loopSpec  `json:"loop"`
}

type hotPath struct {
	ID         string   `json:"id"`
	Category   string   `json:"category"`
	Risk       string   `json:"risk"`
	Files      []string `json:"files"`
	Benchmarks []string `json:"benchmarks,omitempty"`
	Tests      []string `json:"tests,omitempty"`
	Candidate  string   `json:"candidate"`
}

type loopSpec struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
