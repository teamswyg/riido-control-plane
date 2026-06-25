package main

type evidence struct {
	SchemaVersion         string            `json:"schema_version"`
	Status                string            `json:"status"`
	SurfaceCount          int               `json:"surface_count"`
	CandidateCount        int               `json:"candidate_count"`
	BenchmarkCommand      string            `json:"benchmark_command"`
	LocalPressureCommand  string            `json:"local_pressure_command"`
	ManualPressureCommand string            `json:"manual_pressure_command"`
	RaceCommand           string            `json:"race_command"`
	PprofCommands         []string          `json:"pprof_commands"`
	Surfaces              []surfaceEvidence `json:"surfaces"`
}

type surfaceEvidence struct {
	ID        string         `json:"id"`
	Category  string         `json:"category"`
	Risk      string         `json:"risk"`
	Files     []fileEvidence `json:"files"`
	Candidate string         `json:"candidate"`
}

type fileEvidence struct {
	Path    string         `json:"path"`
	Signals map[string]int `json:"signals"`
}

func newEvidence(root string, m manifest) (evidence, error) {
	rows, err := scanSurfaces(root, m.Surfaces)
	if err != nil {
		return evidence{}, err
	}
	return evidence{
		SchemaVersion:         evidenceSchema,
		Status:                "verified",
		SurfaceCount:          len(rows),
		CandidateCount:        len(rows),
		BenchmarkCommand:      m.BenchmarkCommand,
		LocalPressureCommand:  m.LocalPressureCommand,
		ManualPressureCommand: m.ManualPressureCommand,
		RaceCommand:           m.RaceCommand,
		PprofCommands:         m.PprofCommands,
		Surfaces:              rows,
	}, nil
}
