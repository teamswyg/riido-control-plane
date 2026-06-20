package main

type publicConfigKeyMinimization struct {
	RiidoTask              string   `json:"riido_task"`
	CanonicalOwner         string   `json:"canonical_owner"`
	InfraAwarenessOwner    string   `json:"infra_awareness_owner"`
	Rule                   string   `json:"rule"`
	RequiredSecretKeys     []string `json:"required_secret_keys"`
	RequiredVariableKeys   []string `json:"required_variable_keys"`
	OptionalVariableKeys   []string `json:"optional_variable_keys"`
	StableInfraSourceNames []string `json:"stable_infra_source_names"`
	PublicDocsMayReference []string `json:"public_docs_may_reference"`
	PublicDocsMustNotRef   []string `json:"public_docs_must_not_reference"`
	WorkflowKeySource      string   `json:"workflow_key_source"`
}

type publicSensitiveSurfaceGuard struct {
	RiidoTask                  string            `json:"riido_task"`
	CanonicalOwner             string            `json:"canonical_owner"`
	InfraAwarenessOwner        string            `json:"infra_awareness_owner"`
	Rule                       string            `json:"rule"`
	PublicKeyNamesAreSensitive bool              `json:"public_key_names_are_sensitive"`
	KeyNameScopePaths          []string          `json:"key_name_scope_paths"`
	CanonicalCDKeyListPaths    []string          `json:"canonical_cd_key_list_paths"`
	BroadSummaryDocsMustLink   []string          `json:"broad_summary_docs_must_link_not_list_cd_keys"`
	AllowedPublicInformation   []string          `json:"allowed_public_information"`
	AllowedPublicNonCDRuntime  []nonCDRuntimeKey `json:"allowed_public_non_cd_runtime_keys"`
	ForbiddenPublicInformation []string          `json:"forbidden_public_information"`
	InfraMustKnow              []string          `json:"infra_must_know"`
}

type nonCDRuntimeKey struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}
