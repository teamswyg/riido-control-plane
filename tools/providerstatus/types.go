package main

type manifest struct {
	SchemaVersion        string        `json:"schema_version"`
	ID                   string        `json:"id"`
	Title                string        `json:"title"`
	RiidoTask            string        `json:"riido_task"`
	GeneratedDoc         string        `json:"generated_doc"`
	Workflow             string        `json:"workflow"`
	EvidenceArtifact     string        `json:"evidence_artifact"`
	OwnerPackage         string        `json:"owner_package"`
	Surfaces             []surface     `json:"surfaces"`
	RoutingStatuses      []value       `json:"routing_statuses"`
	DistributionChannels []value       `json:"distribution_channels"`
	ValidationRules      []rule        `json:"validation_rules"`
	RoutingRules         []rule        `json:"routing_rules"`
	Authorization        []authRule    `json:"authorization"`
	SourceChecks         []sourceCheck `json:"source_checks"`
	Loop                 evidenceLoop  `json:"loop"`
	NonGoals             []string      `json:"non_goals"`
}

type surface struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type value struct {
	Value string `json:"value"`
	Owner string `json:"owner"`
}

type rule struct {
	ID   string `json:"id"`
	Kind string `json:"kind,omitempty"`
	Rule string `json:"rule"`
}

type authRule struct {
	Action string `json:"action"`
	Scope  string `json:"scope"`
}
