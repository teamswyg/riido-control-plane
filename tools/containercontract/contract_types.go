package main

type imageContract struct {
	SchemaVersion           string        `json:"schema_version"`
	ID                      string        `json:"id"`
	Service                 string        `json:"service"`
	Dockerfile              string        `json:"dockerfile"`
	FargateTaskDefinitionIR string        `json:"fargate_task_definition_ir,omitempty"`
	Assertions              []string      `json:"assertions"`
	Loop                    evidenceLoop  `json:"loop"`
	Build                   buildContract `json:"build"`
	Final                   finalContract `json:"final"`
}

type buildContract struct {
	BuildArg   buildArgContract `json:"build_arg"`
	StageName  string           `json:"stage_name"`
	Workdir    string           `json:"workdir"`
	CGOEnabled string           `json:"cgo_enabled"`
	GoBuild    goBuildContract  `json:"go_build"`
}

type buildArgContract struct {
	Name    string `json:"name"`
	Default string `json:"default"`
}

type goBuildContract struct {
	Package  string   `json:"package"`
	Output   string   `json:"output"`
	Trimpath bool     `json:"trimpath"`
	LDFlags  []string `json:"ldflags"`
}

type finalContract struct {
	BaseImage      string                 `json:"base_image"`
	CopyFrom       string                 `json:"copy_from"`
	CopySource     string                 `json:"copy_source"`
	Binary         string                 `json:"binary"`
	RequiredCopies []requiredCopyContract `json:"required_copies,omitempty"`
	ExposedPorts   []int                  `json:"exposed_ports"`
	Env            map[string]string      `json:"env"`
	User           string                 `json:"user"`
	Entrypoint     []string               `json:"entrypoint"`
}

type requiredCopyContract struct {
	From        string `json:"from"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
}
