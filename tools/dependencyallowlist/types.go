package main

const schemaVersion = "riido-go-dependency-allowlist.v2"

type contract struct {
	SchemaVersion        string          `json:"schema_version"`
	ID                   string          `json:"id"`
	Service              string          `json:"service"`
	Policy               string          `json:"policy"`
	Assertions           []string        `json:"assertions"`
	Loop                 evidenceLoop    `json:"loop"`
	AllowedDirectModules []allowedModule `json:"allowed_direct_modules"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}

type allowedModule struct {
	Path     string `json:"path"`
	Layer    string `json:"layer"`
	Owner    string `json:"owner"`
	Approved bool   `json:"approved"`
	Reason   string `json:"reason"`
}

type goModule struct {
	Path     string
	Version  string
	Main     bool
	Indirect bool
}
