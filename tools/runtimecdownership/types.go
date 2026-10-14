package main

type manifest struct {
	SchemaVersion                       string                        `json:"schema_version"`
	ID                                  string                        `json:"id"`
	RiidoTask                           string                        `json:"riido_task"`
	GeneratedDoc                        string                        `json:"generated_doc"`
	Workflow                            string                        `json:"workflow"`
	EvidenceArtifact                    string                        `json:"evidence_artifact"`
	Hardening                           []string                      `json:"hardening_tasks"`
	Supersedes                          []string                      `json:"supersedes_tasks"`
	Runtime                             string                        `json:"runtime_service"`
	Current                             currentStrategy               `json:"current_strategy"`
	OptionalModes                       []optionalWorkflowMode        `json:"optional_workflow_modes"`
	Future                              []futureStrategy              `json:"future_strategies"`
	Redaction                           redactionPolicy               `json:"public_redaction_policy"`
	Handoff                             handoffPolicy                 `json:"live_handoff_policy"`
	Infra                               infraConsumes                 `json:"infra_consumes"`
	InfraVisibility                     infraVisibilityPolicy         `json:"infra_visibility_policy"`
	PublicExport                        publicExportContract          `json:"public_export_contract"`
	PublicSurfaceScan                   publicSurfaceScanContract     `json:"public_surface_scan_contract"`
	PublicConfigKeyMinimization         publicConfigKeyMinimization   `json:"public_config_key_minimization"`
	PublicSensitiveSurfaceGuard         publicSensitiveSurfaceGuard   `json:"public_sensitive_surface_guard"`
	PublicOperationalDetailMinimization operationalDetailMinimization `json:"public_operational_detail_minimization"`
	CodeDeployActivationGate            codeDeployActivationGate      `json:"codedeploy_activation_gate"`
	InfraTopology                       infraTopologyContract         `json:"infra_topology_contract"`
	DependencyDirection                 dependencyDirection           `json:"dependency_direction"`
	Loop                                evidenceLoop                  `json:"loop"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}

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
