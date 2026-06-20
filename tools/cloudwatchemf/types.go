package main

type manifest struct {
	SchemaVersion      string         `json:"schema_version"`
	ID                 string         `json:"id"`
	Title              string         `json:"title"`
	GeneratedDoc       string         `json:"generated_doc"`
	Workflow           string         `json:"workflow"`
	EvidenceArtifact   string         `json:"evidence_artifact"`
	OwnerPackage       string         `json:"owner_package"`
	SourceChecks       []sourceCheck  `json:"source_checks"`
	RequiredDimensions []string       `json:"required_dimensions"`
	RequiredJSONFields []string       `json:"required_json_fields"`
	RequiredMetricUnit []requiredUnit `json:"required_metric_units"`
	Loop               evidenceLoop   `json:"loop"`
}

type sourceCheck struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Contains []string `json:"contains"`
}

type requiredUnit struct {
	Name string `json:"name"`
	Unit string `json:"unit"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}

type options struct {
	Repo        string
	Manifest    string
	EvidenceOut string
	WriteDoc    bool
	CheckDoc    bool
}
