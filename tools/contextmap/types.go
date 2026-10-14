package main

type manifest struct {
	SchemaVersion    string            `json:"schema_version"`
	ID               string            `json:"id"`
	Title            string            `json:"title"`
	RiidoTask        string            `json:"riido_task"`
	GeneratedDoc     string            `json:"generated_doc"`
	Workflow         string            `json:"workflow"`
	EvidenceArtifact string            `json:"evidence_artifact"`
	OwnedContexts    []ownedContext    `json:"owned_contexts"`
	ImportedContexts []importedContext `json:"imported_contexts"`
	ExternalContexts []externalContext `json:"external_contexts"`
	DirectionRules   directionRules    `json:"direction_rules"`
	SSOTLinks        []link            `json:"ssot_links"`
	SourceChecks     []sourceCheck     `json:"source_checks"`
	Loop             evidenceLoop      `json:"loop"`
}

type ownedContext struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	OwnerPaths     []string `json:"owner_paths"`
	Responsibility string   `json:"responsibility"`
}

type importedContext struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ImportedFrom string `json:"imported_from"`
	Use          string `json:"use"`
}

type externalContext struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Owner    string `json:"owner"`
	Boundary string `json:"boundary"`
}

type directionRules struct {
	AllowedImports      []string `json:"allowed_imports"`
	ForbiddenGoImports  []string `json:"forbidden_go_imports"`
	ForbiddenPathTokens []string `json:"forbidden_path_tokens"`
}
