package main

const schemaVersion = "riido-go-dependency-allowlist.v2"

type contract struct {
	SchemaVersion        string          `json:"schema_version"`
	Service              string          `json:"service"`
	Policy               string          `json:"policy"`
	AllowedDirectModules []allowedModule `json:"allowed_direct_modules"`
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
