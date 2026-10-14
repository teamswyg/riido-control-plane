package main

type manifest struct {
	SchemaVersion    string        `json:"schema_version"`
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	RiidoTask        string        `json:"riido_task"`
	GeneratedDoc     string        `json:"generated_doc"`
	Workflow         string        `json:"workflow"`
	EvidenceArtifact string        `json:"evidence_artifact"`
	OwnerPackage     string        `json:"owner_package"`
	Ports            []surface     `json:"ports"`
	Records          []surface     `json:"records"`
	ReplayRules      []rule        `json:"replay_rules"`
	VersionConstants []constant    `json:"version_constants"`
	SourceChecks     []sourceCheck `json:"source_checks"`
	Loop             evidenceLoop  `json:"loop"`
	NonGoals         []string      `json:"non_goals"`
}

type surface struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type rule struct {
	ID   string `json:"id"`
	Kind string `json:"kind,omitempty"`
	Rule string `json:"rule"`
}

type sourceCheck struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Contains []string `json:"contains"`
}

type constant struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
