package main

type manifest struct {
	SchemaVersion     string       `json:"schema_version"`
	ID                string       `json:"id"`
	Title             string       `json:"title"`
	RiidoTask         string       `json:"riido_task"`
	GeneratedDoc      string       `json:"generated_doc"`
	Workflow          string       `json:"workflow"`
	EvidenceArtifact  string       `json:"evidence_artifact"`
	LoopSource        string       `json:"loop_source"`
	ModulePath        string       `json:"module_path"`
	License           string       `json:"license"`
	Fragments         []string     `json:"fragments"`
	Summary           []string     `json:"summary"`
	Owns              []string     `json:"owns"`
	DoesNotOwn        []string     `json:"does_not_own"`
	Rationale         []string     `json:"rationale"`
	DocLinks          []docLink    `json:"doc_links"`
	Development       development  `json:"development"`
	RuntimeCD         runtimeCD    `json:"runtime_cd"`
	ContractFlow      []string     `json:"contract_flow"`
	Verification      []string     `json:"verification"`
	Rules             []string     `json:"rules"`
	RequiredMarkers   []string     `json:"required_markers"`
	ForbiddenLiterals []string     `json:"forbidden_literals"`
	Loop              evidenceLoop `json:"loop"`
}

type docLink struct {
	Topic string `json:"topic"`
	Path  string `json:"path"`
}

type development struct {
	Intro     []string   `json:"intro"`
	Env       []envEntry `json:"env"`
	Workflows []string   `json:"workflows"`
	Endpoints []string   `json:"endpoints"`
}

type envEntry struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
