package main

type manifest struct {
	SchemaVersion       string        `json:"schema_version"`
	ID                  string        `json:"id"`
	Title               string        `json:"title"`
	RiidoTask           string        `json:"riido_task"`
	GeneratedDoc        string        `json:"generated_doc"`
	Workflow            string        `json:"workflow"`
	EvidenceArtifact    string        `json:"evidence_artifact"`
	EvidenceProfiles    []profile     `json:"evidence_profiles"`
	OwnerPackage        string        `json:"owner_package"`
	Roles               []string      `json:"roles"`
	Visibilities        []string      `json:"visibilities"`
	Actions             []string      `json:"actions"`
	VisibilityRules     []rule        `json:"visibility_rules"`
	AuthorizationScopes []string      `json:"authorization_scopes"`
	Routes              []string      `json:"routes"`
	RequestDTOs         []string      `json:"request_dtos"`
	ResponseDTOs        []string      `json:"response_dtos"`
	StoreMethods        []string      `json:"store_methods"`
	SourceChecks        []sourceCheck `json:"source_checks"`
	NonGoals            []string      `json:"non_goals"`
	Loop                evidenceLoop  `json:"loop"`
}

type rule struct {
	ID             string   `json:"id"`
	Subject        string   `json:"subject"`
	Record         string   `json:"record"`
	AllowedActions []string `json:"allowed_actions,omitempty"`
	DeniedActions  []string `json:"denied_actions,omitempty"`
	Reason         string   `json:"reason"`
}

type profile struct {
	ID               string   `json:"id"`
	Workflow         string   `json:"workflow"`
	EvidenceArtifact string   `json:"evidence_artifact"`
	Focus            string   `json:"focus"`
	TestPatterns     []string `json:"test_patterns"`
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
