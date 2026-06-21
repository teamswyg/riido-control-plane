package main

type manifest struct {
	SchemaVersion     string             `json:"schema_version"`
	ID                string             `json:"id"`
	Title             string             `json:"title"`
	GeneratedDoc      string             `json:"generated_doc"`
	Workflow          string             `json:"workflow"`
	EvidenceArtifact  string             `json:"evidence_artifact"`
	ScanRoots         []string           `json:"scan_roots"`
	ScanFiles         []string           `json:"scan_files"`
	Standalone        []standalone       `json:"standalone_manifests"`
	SourceManifests   []sourceSSOT       `json:"source_manifests"`
	ContractArtifacts []contractArtifact `json:"contract_artifacts"`
	ImportedManifests []importedManifest `json:"imported_manifests"`
	OwnedManifests    []ownedManifest    `json:"owned_manifests"`
	ManualGroups      []manualGroup      `json:"manual_groups"`
	Assertions        []string           `json:"assertions"`
	Loop              evidenceLoop       `json:"loop"`
}

type standalone struct {
	Path             string `json:"path"`
	EvidenceTool     string `json:"evidence_tool"`
	Workflow         string `json:"workflow"`
	EvidenceArtifact string `json:"evidence_artifact"`
	HumanDoc         string `json:"human_doc,omitempty"`
}

type manualGroup struct {
	ID           string   `json:"id"`
	Owner        string   `json:"owner"`
	Reason       string   `json:"reason"`
	NextArtifact string   `json:"next_artifact"`
	Paths        []string `json:"paths,omitempty"`
	PathPrefixes []string `json:"path_prefixes,omitempty"`
}

type docClass struct {
	Path          string
	Kind          string
	Group         string
	Reason        string
	GeneratorTool string
	EvidenceTool  string
	HasLoop       bool
}

type manualDir struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

type manifestDir struct {
	Group string `json:"group"`
	Count int    `json:"count"`
}

type manualSample struct {
	Group  string `json:"group"`
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
