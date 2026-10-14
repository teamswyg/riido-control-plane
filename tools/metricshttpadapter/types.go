package main

type manifest struct {
	SchemaVersion    string           `json:"schema_version"`
	ID               string           `json:"id"`
	Title            string           `json:"title"`
	GeneratedDoc     string           `json:"generated_doc"`
	Workflow         string           `json:"workflow"`
	EvidenceArtifact string           `json:"evidence_artifact"`
	OwnerPackage     string           `json:"owner_package"`
	Endpoint         endpointContract `json:"endpoint"`
	SourceChecks     []sourceCheck    `json:"source_checks"`
	RequiredFields   []string         `json:"required_json_fields"`
	RequiredStatuses []statusContract `json:"required_statuses"`
	Loop             evidenceLoop     `json:"loop"`
}

type endpointContract struct {
	Method   string `json:"method"`
	Path     string `json:"path"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type sourceCheck struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Contains []string `json:"contains"`
}

type statusContract struct {
	Case   string `json:"case"`
	Status int    `json:"status"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
