package main

type childNativeMapping struct {
	Checkout            checkoutMapping `json:"checkout"`
	GoToolchain         goMapping       `json:"go_toolchain"`
	RepositoryReadme    commandMapping  `json:"repository_readme,omitempty"`
	ExecutableKnowledge commandMapping  `json:"executable_knowledge,omitempty"`
	DependencyAllowlist commandMapping  `json:"dependency_allowlist,omitempty"`
	ContextMap          commandMapping  `json:"context_map,omitempty"`
	ModuleDecomposition commandMapping  `json:"module_decomposition,omitempty"`
	ModuleDownload      commandMapping  `json:"module_download,omitempty"`
	GoCIBaseline        commandMapping  `json:"go_ci_baseline,omitempty"`
	LintInstall         commandMapping  `json:"lint_install,omitempty"`
	Lint                commandMapping  `json:"lint,omitempty"`
	Coverage            commandMapping  `json:"coverage,omitempty"`
	EvidenceArtifact    artifactMapping `json:"evidence_artifact"`
}

type childParityClaim struct {
	AllSourceStepsMapped              bool `json:"all_source_steps_mapped"`
	RequiredAdapterCount              int  `json:"required_adapter_count"`
	RepositoryReadmeCommandExact      bool `json:"repository_readme_command_exact,omitempty"`
	ExecutableKnowledgeCommandExact   bool `json:"executable_knowledge_command_exact,omitempty"`
	DependencyAllowlistBehaviorExact  bool `json:"dependency_allowlist_behavior_exact,omitempty"`
	ContextMapCommandExact            bool `json:"context_map_command_exact,omitempty"`
	ModuleDecompositionCommandExact   bool `json:"module_decomposition_command_exact,omitempty"`
	ModuleDownloadCommandExact        bool `json:"module_download_command_exact,omitempty"`
	GoCIBaselineCommandExact          bool `json:"go_ci_baseline_command_exact,omitempty"`
	LintInstallCommandExact           bool `json:"lint_install_command_exact,omitempty"`
	LintCommandExact                  bool `json:"lint_command_exact,omitempty"`
	CoverageCommandExact              bool `json:"coverage_command_exact,omitempty"`
	CoverageSummaryBehaviorExact      bool `json:"coverage_summary_behavior_exact,omitempty"`
	SecureEvidencePermissions         bool `json:"secure_evidence_permissions"`
	SourceWorkflowEdited              bool `json:"source_workflow_edited"`
	SourceWorkflowExecutedByThisSlice bool `json:"source_workflow_executed_by_this_slice"`
}
