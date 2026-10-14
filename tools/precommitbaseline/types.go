package main

type manifest struct {
	SchemaVersion   string       `json:"schema_version"`
	ID              string       `json:"id"`
	Title           string       `json:"title"`
	GeneratedDoc    string       `json:"generated_doc"`
	Workflow        string       `json:"workflow"`
	Evidence        string       `json:"evidence_artifact"`
	EvidenceTTL     int          `json:"evidence_ttl_hours"`
	PreCommitConfig string       `json:"pre_commit_config"`
	Hooks           []checkBlock `json:"hooks"`
	Scripts         []scriptSpec `json:"scripts"`
	Loop            loopRecord   `json:"loop"`
}

type checkBlock struct {
	ID       string   `json:"id"`
	Summary  string   `json:"summary"`
	Contains []string `json:"contains"`
}

type scriptSpec struct {
	Path     string   `json:"path"`
	Summary  string   `json:"summary"`
	Contains []string `json:"contains"`
}
