package main

type openAPISpec struct {
	Paths map[string]map[string]operation `json:"paths"`
}

type operation struct {
	OperationID    string         `json:"operationId"`
	Summary        string         `json:"summary"`
	Deprecated     bool           `json:"deprecated"`
	Client         clientMetadata `json:"x-riido-client"`
	Lifecycle      string         `json:"x-riido-lifecycle"`
	Replacement    string         `json:"x-riido-replacement"`
	RemovalHorizon string         `json:"x-riido-removal-horizon"`
}

type clientMetadata struct {
	GeneratedPath string `json:"generated_path"`
}

type operationRow struct {
	Method         string
	Path           string
	OperationID    string
	Summary        string
	GeneratedPath  string
	Deprecated     bool
	Lifecycle      string
	Replacement    string
	RemovalHorizon string
}

type config struct {
	OpenAPI          string
	DSL              string
	IR               string
	Core             string
	React            string
	Out              string
	PRBody           string
	PreviousManifest string
	SourceRepo       string
	SourceRef        string
	SourceCommit     string
	TargetRepo       string
	TargetBranch     string
	GeneratedAt      string
}

type previousManifest struct {
	Available    bool
	SourceCommit string
	Operations   []operationRow
}

type changeSection struct {
	Title string
	Lines []string
}
