package main

type manifest struct {
	SchemaVersion            string        `json:"schema_version"`
	ID                       string        `json:"id"`
	Title                    string        `json:"title"`
	RiidoTask                string        `json:"riido_task"`
	GeneratedDoc             string        `json:"generated_doc"`
	Workflow                 string        `json:"workflow"`
	EvidenceArtifact         string        `json:"evidence_artifact"`
	OwnerPackage             string        `json:"owner_package"`
	Surfaces                 []surface     `json:"surfaces"`
	Resources                []string      `json:"resources"`
	TokenTransports          []transport   `json:"token_transports"`
	RuntimeConfigKeys        []string      `json:"runtime_config_keys"`
	ExternalContractVersions []string      `json:"external_contract_versions"`
	RuleGroups               []ruleGroup   `json:"rule_groups"`
	SourceChecks             []sourceCheck `json:"source_checks"`
	Loop                     loop          `json:"loop"`
	NonGoals                 []string      `json:"non_goals"`
}

type surface struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type transport struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

type ruleGroup struct {
	ID    string   `json:"id"`
	Rules []string `json:"rules"`
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
