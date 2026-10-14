package main

type manifest struct {
	SchemaVersion    string        `json:"schema_version"`
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	GeneratedDoc     string        `json:"generated_doc"`
	Workflow         string        `json:"workflow"`
	EvidenceArtifact string        `json:"evidence_artifact"`
	DomainSSOT       string        `json:"domain_ssot"`
	OwnerPackage     string        `json:"owner_package"`
	Cases            []routingCase `json:"cases"`
	SourceChecks     []sourceCheck `json:"source_checks"`
	Loop             evidenceLoop  `json:"loop"`
}

type routingCase struct {
	Name              string        `json:"name"`
	RuntimeProvider   string        `json:"runtime_provider"`
	ProviderStatuses  []providerRow `json:"provider_statuses"`
	WantAllowed       bool          `json:"want_allowed"`
	WantRoutingStatus string        `json:"want_routing_status"`
	WantReason        string        `json:"want_reason"`
	WantErrorContains string        `json:"want_error_contains"`
}

type providerRow struct {
	ProviderKind  string `json:"provider_kind"`
	RoutingStatus string `json:"routing_status"`
}

type sourceCheck struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Contains []string `json:"contains"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
