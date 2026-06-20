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
	Packages         []packageEntry `json:"packages"`
	Rules            []string       `json:"rules"`
	Loop             evidenceLoop   `json:"loop"`
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
