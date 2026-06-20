package main

type checkRecord struct {
	SchemaVersion           string       `json:"schema_version"`
	ID                      string       `json:"id"`
	Service                 string       `json:"service"`
	Status                  string       `json:"status"`
	Dockerfile              string       `json:"dockerfile"`
	FargateTaskDefinitionIR string       `json:"fargate_task_definition_ir,omitempty"`
	BuildStage              string       `json:"build_stage"`
	FinalBaseImage          string       `json:"final_base_image"`
	FinalUser               string       `json:"final_user"`
	Entrypoint              []string     `json:"entrypoint"`
	ExposedPorts            []int        `json:"exposed_ports"`
	Loop                    evidenceLoop `json:"loop"`
	ChecksTotal             int          `json:"checks_total"`
}
