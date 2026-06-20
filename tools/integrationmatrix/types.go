package main

type manifest struct {
	SchemaVersion string        `json:"schema_version"`
	ID            string        `json:"id"`
	Title         string        `json:"title"`
	RiidoTask     string        `json:"riido_task"`
	GeneratedDoc  string        `json:"generated_doc"`
	Workflow      string        `json:"workflow"`
	Evidence      string        `json:"evidence_artifact"`
	PublicGates   []publicGate  `json:"public_gates"`
	PrivateGates  []privateGate `json:"private_gates"`
	Rules         []string      `json:"rules"`
	ForbiddenDeps []string      `json:"forbidden_pull_request_dependencies"`
	Loop          evidenceLoop  `json:"loop"`
}

type publicGate struct {
	Surface            string   `json:"surface"`
	Verification       string   `json:"verification"`
	ExternalDependency string   `json:"external_dependency"`
	PullRequestGate    bool     `json:"pull_request_gate"`
	Workflows          []string `json:"workflows"`
	Commands           []string `json:"commands"`
}

type privateGate struct {
	Surface  string `json:"surface"`
	Owner    string `json:"owner"`
	Evidence string `json:"evidence"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
