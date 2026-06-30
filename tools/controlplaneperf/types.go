package main

type manifest struct {
	SchemaVersion             string                  `json:"schema_version"`
	ID                        string                  `json:"id"`
	Title                     string                  `json:"title"`
	GeneratedDoc              string                  `json:"generated_doc"`
	Workflow                  string                  `json:"workflow"`
	EvidenceArtifact          string                  `json:"evidence_artifact"`
	BenchmarkArtifact         string                  `json:"benchmark_artifact"`
	RaceArtifact              string                  `json:"race_artifact"`
	LocalPressureArtifact     string                  `json:"local_pressure_artifact"`
	SummaryArtifact           string                  `json:"summary_artifact"`
	CandidateArtifact         string                  `json:"candidate_artifact"`
	PressureCandidateArtifact string                  `json:"pressure_candidate_artifact"`
	ArchitectureQueryArtifact string                  `json:"architecture_query_artifact"`
	BenchmarkHistory          string                  `json:"benchmark_history"`
	EvidenceTool              string                  `json:"evidence_tool"`
	BenchmarkCommand          string                  `json:"benchmark_command"`
	BenchmarkHistoryCommand   string                  `json:"benchmark_history_command"`
	LocalPressureCommand      string                  `json:"local_pressure_command"`
	ManualPressureCommand     string                  `json:"manual_pressure_command"`
	LocalPprofCommand         string                  `json:"local_pprof_command"`
	ArchitectureQueryCommand  string                  `json:"architecture_query_command"`
	RaceCommand               string                  `json:"race_command"`
	PprofCommand              string                  `json:"pprof_command"`
	LiveLoadCommand           string                  `json:"live_load_command"`
	LocalPressureScenarios    []string                `json:"local_pressure_scenarios"`
	Sources                   []pressureSource        `json:"sources"`
	ArchitectureComponents    []architectureComponent `json:"architecture_components"`
	HotPaths                  []hotPath               `json:"hot_paths"`
	Assertions                []string                `json:"assertions"`
	Loop                      loopSpec                `json:"loop"`
}

type pressureSource struct {
	ID                    string   `json:"id"`
	SourceWorkflow        string   `json:"source_workflow"`
	SummaryArtifact       string   `json:"summary_artifact"`
	CandidateArtifact     string   `json:"candidate_artifact"`
	HarnessLoop           string   `json:"harness_loop"`
	PromotionTarget       string   `json:"promotion_target"`
	RequiredNextArtifacts []string `json:"required_next_artifacts"`
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
