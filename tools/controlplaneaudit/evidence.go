package main

type evidence struct {
	SchemaVersion         string            `json:"schema_version"`
	Status                string            `json:"status"`
	SurfaceCount          int               `json:"surface_count"`
	CandidateCount        int               `json:"candidate_count"`
	AssertionCount        int               `json:"assertion_count"`
	RaceArtifact          string            `json:"race_artifact"`
	BenchmarkCommand      string            `json:"benchmark_command"`
	LocalPressureCommand  string            `json:"local_pressure_command"`
	ManualPressureCommand string            `json:"manual_pressure_command"`
	RaceCommand           string            `json:"race_command"`
	PprofCommands         []string          `json:"pprof_commands"`
	RequiredCategories    []string          `json:"required_categories"`
	MissingCategories     []string          `json:"missing_categories"`
	Assertions            []string          `json:"assertions"`
	CategoryCounts        map[string]int    `json:"category_counts"`
	Surfaces              []surfaceEvidence `json:"surfaces"`
	Loop                  loopSpec          `json:"loop"`
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
	counts := categoryCounts(rows)
	return evidence{
		SchemaVersion:         evidenceSchema,
		Status:                "verified",
		SurfaceCount:          len(rows),
		CandidateCount:        len(rows),
		AssertionCount:        len(m.Assertions),
		RaceArtifact:          m.RaceArtifact,
		BenchmarkCommand:      m.BenchmarkCommand,
		LocalPressureCommand:  m.LocalPressureCommand,
		ManualPressureCommand: m.ManualPressureCommand,
		RaceCommand:           m.RaceCommand,
		PprofCommands:         m.PprofCommands,
		RequiredCategories:    append([]string(nil), m.RequiredCategories...),
		MissingCategories:     missingCategories(m.RequiredCategories, counts),
		Assertions:            append([]string(nil), m.Assertions...),
		CategoryCounts:        counts,
		Surfaces:              rows,
		Loop:                  m.Loop,
	}, nil
}
