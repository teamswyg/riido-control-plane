package deploypolicy

type runtimeCDOwnership struct {
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
