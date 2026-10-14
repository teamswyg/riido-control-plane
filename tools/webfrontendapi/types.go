package main

type manifest struct {
	SchemaVersion     string        `json:"schema_version"`
	ID                string        `json:"id"`
	Title             string        `json:"title"`
	GeneratedDoc      string        `json:"generated_doc"`
	Workflow          string        `json:"workflow"`
	EvidenceArtifact  string        `json:"evidence_artifact"`
	OwnerPackages     []string      `json:"owner_packages"`
	RuntimeConfigKeys []string      `json:"runtime_config_keys"`
	CORSCases         []corsCase    `json:"cors_cases"`
	SourceChecks      []sourceCheck `json:"source_checks"`
	Loop              evidenceLoop  `json:"loop"`
}

type corsCase struct {
	Name            string   `json:"name"`
	AllowedOrigins  []string `json:"allowed_origins"`
	Method          string   `json:"method"`
	Path            string   `json:"path"`
	Origin          string   `json:"origin"`
	RequestMethod   string   `json:"request_method"`
	RequestHeaders  string   `json:"request_headers"`
	WantStatus      int      `json:"want_status"`
	WantAllowOrigin string   `json:"want_allow_origin"`
	WantCredentials string   `json:"want_credentials"`
}

type sourceCheck struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Contains []string `json:"contains"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
