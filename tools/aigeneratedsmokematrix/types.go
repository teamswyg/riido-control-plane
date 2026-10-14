package main

type manifest struct {
	SchemaVersion         string          `json:"schema_version"`
	ID                    string          `json:"id"`
	Title                 string          `json:"title"`
	RiidoTask             string          `json:"riido_task"`
	GeneratedDoc          string          `json:"generated_doc"`
	Workflow              string          `json:"workflow"`
	EvidenceArtifact      string          `json:"evidence_artifact"`
	OpenAPI               string          `json:"openapi"`
	SmokeMatrix           string          `json:"smoke_matrix"`
	SmokeSchemaVersion    string          `json:"smoke_schema_version"`
	OperationCounts       operationCounts `json:"operation_counts"`
	RequiredEvidenceTests []string        `json:"required_evidence_tests"`
	SourceChecks          []sourceCheck   `json:"source_checks"`
	Loop                  evidenceLoop    `json:"loop"`
	NonGoals              []string        `json:"non_goals"`
}

type operationCounts struct {
	Total int `json:"total"`
	V1    int `json:"v1"`
	V2    int `json:"v2"`
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
