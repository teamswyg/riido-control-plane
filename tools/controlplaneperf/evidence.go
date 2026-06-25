package main

type evidence struct {
	SchemaVersion         string              `json:"schema_version"`
	Status                string              `json:"status"`
	HotPathCount          int                 `json:"hot_path_count"`
	BenchmarkCount        int                 `json:"benchmark_count"`
	TestCount             int                 `json:"test_count"`
	CandidateCount        int                 `json:"candidate_count"`
	BenchmarkCommand      string              `json:"benchmark_command"`
	LocalPressureCommand  string              `json:"local_pressure_command"`
	ManualPressureCommand string              `json:"manual_pressure_command"`
	RaceCommand           string              `json:"race_command"`
	PprofCommand          string              `json:"pprof_command"`
	LiveLoadCommand       string              `json:"live_load_command"`
	HotPaths              []hotPathEvidence   `json:"hot_paths"`
	Candidates            []candidateEvidence `json:"candidates"`
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
	return evidence{
		SchemaVersion:         evidenceSchema,
		Status:                "verified",
		HotPathCount:          len(m.HotPaths),
		BenchmarkCount:        benchmarkCount(m.HotPaths),
		TestCount:             testCount(m.HotPaths),
		CandidateCount:        len(m.HotPaths),
		BenchmarkCommand:      m.BenchmarkCommand,
		LocalPressureCommand:  m.LocalPressureCommand,
		ManualPressureCommand: m.ManualPressureCommand,
		RaceCommand:           m.RaceCommand,
		PprofCommand:          m.PprofCommand,
		LiveLoadCommand:       m.LiveLoadCommand,
		HotPaths:              hotPathEvidenceRows(m.HotPaths),
		Candidates:            candidateEvidenceRows(m.HotPaths),
	}
}
