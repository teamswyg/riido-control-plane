package main

type evidence struct {
	SchemaVersion                  string                          `json:"schema_version"`
	Status                         string                          `json:"status"`
	GeneratedAt                    string                          `json:"generated_at"`
	ExpiresAt                      string                          `json:"expires_at"`
	HotPathCount                   int                             `json:"hot_path_count"`
	BenchmarkCount                 int                             `json:"benchmark_count"`
	TestCount                      int                             `json:"test_count"`
	CandidateCount                 int                             `json:"candidate_count"`
	ArchitectureComponentCount     int                             `json:"architecture_component_count"`
	ArchitectureFileCount          int                             `json:"architecture_file_count"`
	ArchitectureTargetCommandCount int                             `json:"architecture_target_command_count"`
	AssertionCount                 int                             `json:"assertion_count"`
	BenchmarkCommand               string                          `json:"benchmark_command"`
	RaceArtifact                   string                          `json:"race_artifact"`
	LocalPressureCommand           string                          `json:"local_pressure_command"`
	PressureCandidateArtifact      string                          `json:"pressure_candidate_artifact"`
	ArchitectureQueryArtifact      string                          `json:"architecture_query_artifact"`
	ManualPressureCommand          string                          `json:"manual_pressure_command"`
	LocalPprofCommand              string                          `json:"local_pprof_command"`
	ArchitectureQueryCommand       string                          `json:"architecture_query_command"`
	RaceCommand                    string                          `json:"race_command"`
	PprofCommand                   string                          `json:"pprof_command"`
	LiveLoadCommand                string                          `json:"live_load_command"`
	LocalPressureScenarios         []string                        `json:"local_pressure_scenarios"`
	Sources                        []pressureSource                `json:"sources"`
	Assertions                     []string                        `json:"assertions"`
	ArchitectureComponents         []architectureComponentEvidence `json:"architecture_components"`
	FileArchitectureIndex          []architectureFileEvidence      `json:"file_architecture_index"`
	HotPaths                       []hotPathEvidence               `json:"hot_paths"`
	Candidates                     []candidateEvidence             `json:"candidates"`
	Loop                           loopSpec                        `json:"loop"`
}

type hotPathEvidence struct {
	ID         string   `json:"id"`
	Category   string   `json:"category"`
	Files      []string `json:"files"`
	Benchmarks []string `json:"benchmarks,omitempty"`
	Tests      []string `json:"tests,omitempty"`
}

type candidateEvidence struct {
	ID        string `json:"id"`
	Risk      string `json:"risk"`
	Candidate string `json:"candidate"`
}
