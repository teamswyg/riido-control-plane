package main

type manifest struct {
	SchemaVersion    string         `json:"schema_version"`
	ID               string         `json:"id"`
	Title            string         `json:"title"`
	RiidoTask        string         `json:"riido_task"`
	GeneratedDoc     string         `json:"generated_doc"`
	Workflow         string         `json:"workflow"`
	Evidence         string         `json:"evidence_artifact"`
	ModulePath       string         `json:"module_path"`
	SourceRoots      []string       `json:"source_roots"`
	ForbiddenImports []string       `json:"forbidden_imports"`
	FileLineBudget   fileLineBudget `json:"file_line_budget"`
	Packages         []packageEntry `json:"packages"`
	Rules            []string       `json:"rules"`
	Loop             evidenceLoop   `json:"loop"`
}

type fileLineBudget struct {
	TargetLines        int `json:"target_lines"`
	SampleLimit        int `json:"sample_limit"`
	HotspotLimit       int `json:"hotspot_limit"`
	MaxFilesOverTarget int `json:"max_files_over_target"`
	MaxFileLines       int `json:"max_file_lines"`
}

type packageEntry struct {
	Path       string `json:"path"`
	Kind       string `json:"kind"`
	Role       string `json:"role"`
	MustNotOwn string `json:"must_not_own"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
