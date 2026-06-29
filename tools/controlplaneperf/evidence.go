package main

type evidence struct {
	SchemaVersion             string              `json:"schema_version"`
	Status                    string              `json:"status"`
	GeneratedAt               string              `json:"generated_at"`
	ExpiresAt                 string              `json:"expires_at"`
	HotPathCount              int                 `json:"hot_path_count"`
	BenchmarkCount            int                 `json:"benchmark_count"`
	TestCount                 int                 `json:"test_count"`
	CandidateCount            int                 `json:"candidate_count"`
	AssertionCount            int                 `json:"assertion_count"`
	BenchmarkCommand          string              `json:"benchmark_command"`
	RaceArtifact              string              `json:"race_artifact"`
	LocalPressureCommand      string              `json:"local_pressure_command"`
	PressureCandidateArtifact string              `json:"pressure_candidate_artifact"`
	ManualPressureCommand     string              `json:"manual_pressure_command"`
	LocalPprofCommand         string              `json:"local_pprof_command"`
	RaceCommand               string              `json:"race_command"`
	PprofCommand              string              `json:"pprof_command"`
	LiveLoadCommand           string              `json:"live_load_command"`
	LocalPressureScenarios    []string            `json:"local_pressure_scenarios"`
	Sources                   []pressureSource    `json:"sources"`
	Assertions                []string            `json:"assertions"`
	HotPaths                  []hotPathEvidence   `json:"hot_paths"`
	Candidates                []candidateEvidence `json:"candidates"`
	Loop                      loopSpec            `json:"loop"`
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

func newEvidence(m manifest) evidence {
	generatedAt, expiresAt := evidenceWindow(controlPlanePerformanceEvidenceTTLHours)
	return evidence{
		SchemaVersion:             evidenceSchema,
		Status:                    "verified",
		GeneratedAt:               generatedAt,
		ExpiresAt:                 expiresAt,
		HotPathCount:              len(m.HotPaths),
		BenchmarkCount:            benchmarkCount(m.HotPaths),
		TestCount:                 testCount(m.HotPaths),
		CandidateCount:            len(m.HotPaths),
		AssertionCount:            len(m.Assertions),
		BenchmarkCommand:          m.BenchmarkCommand,
		RaceArtifact:              m.RaceArtifact,
		LocalPressureCommand:      m.LocalPressureCommand,
		PressureCandidateArtifact: m.PressureCandidateArtifact,
		ManualPressureCommand:     m.ManualPressureCommand,
		LocalPprofCommand:         m.LocalPprofCommand,
		RaceCommand:               m.RaceCommand,
		PprofCommand:              m.PprofCommand,
		LiveLoadCommand:           m.LiveLoadCommand,
		LocalPressureScenarios:    append([]string(nil), m.LocalPressureScenarios...),
		Sources:                   append([]pressureSource(nil), m.Sources...),
		Assertions:                append([]string(nil), m.Assertions...),
		HotPaths:                  hotPathEvidenceRows(m.HotPaths),
		Candidates:                candidateEvidenceRows(m.HotPaths),
		Loop:                      m.Loop,
	}
}
