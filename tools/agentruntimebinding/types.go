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
	BindingFields    []field       `json:"binding_fields"`
	BindingRules     []rule        `json:"binding_rules"`
	DeviceRules      []rule        `json:"device_rules"`
	SourceChecks     []sourceCheck `json:"source_checks"`
	Loop             evidenceLoop  `json:"loop"`
	NonGoals         []string      `json:"non_goals"`
}

type field struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
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

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
