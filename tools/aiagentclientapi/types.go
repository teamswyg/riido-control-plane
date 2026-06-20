package main

type manifest struct {
	SchemaVersion          string          `json:"schema_version"`
	ID                     string          `json:"id"`
	Title                  string          `json:"title"`
	RiidoTask              string          `json:"riido_task"`
	GeneratedDoc           string          `json:"generated_doc"`
	Workflow               string          `json:"workflow"`
	EvidenceArtifact       string          `json:"evidence_artifact"`
	OpenAPI                string          `json:"openapi"`
	SmokeMatrix            string          `json:"smoke_matrix"`
	DSL                    string          `json:"dsl"`
	IR                     string          `json:"ir"`
	GeneratedCore          string          `json:"generated_core"`
	GeneratedReact         string          `json:"generated_react"`
	OperationCounts        operationCounts `json:"operation_counts"`
	RequiredGeneratedPaths []string        `json:"required_generated_paths"`
	RuntimeConfigKeys      []string        `json:"runtime_config_keys"`
	PublicFields           []string        `json:"public_fields"`
	DeploymentEvidence     []string        `json:"deployment_evidence"`
	SourceChecks           []sourceCheck   `json:"source_checks"`
	Loop                   loop            `json:"loop"`
	NonGoals               []string        `json:"non_goals"`
}

type operationCounts struct {
	Total       int `json:"total"`
	V1          int `json:"v1"`
	V2          int `json:"v2"`
	SmokeMatrix int `json:"smoke_matrix"`
}

type sourceCheck struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Contains []string `json:"contains"`
}

type loop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
