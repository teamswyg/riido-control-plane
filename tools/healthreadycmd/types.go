package main

type manifest struct {
	SchemaVersion    string             `json:"schema_version"`
	ID               string             `json:"id"`
	Title            string             `json:"title"`
	GeneratedDoc     string             `json:"generated_doc"`
	Workflow         string             `json:"workflow"`
	EvidenceArtifact string             `json:"evidence_artifact"`
	OwnerPackages    []string           `json:"owner_packages"`
	Endpoints        []endpointContract `json:"endpoints"`
	SourceChecks     []sourceCheck      `json:"source_checks"`
	CommandTests     []string           `json:"command_tests"`
	Loop             evidenceLoop       `json:"loop"`
}

type endpointContract struct {
	Name       string `json:"name"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Status     string `json:"status"`
	HTTPStatus int    `json:"http_status"`
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
