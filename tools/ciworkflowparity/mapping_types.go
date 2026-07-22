package main

type nativeMapping struct {
	Checkout            checkoutMapping `json:"checkout"`
	GoToolchain         goMapping       `json:"go_toolchain"`
	DependencyAllowlist commandMapping  `json:"dependency_allowlist"`
	EvidenceArtifact    artifactMapping `json:"evidence_artifact"`
	Test                commandMapping  `json:"test"`
}

type checkoutMapping struct {
	Source          string `json:"source"`
	NativeKind      string `json:"native_kind"`
	AdapterRequired bool   `json:"adapter_required"`
}

type goMapping struct {
	Source          string `json:"source"`
	NativeKind      string `json:"native_kind"`
	VersionSource   string `json:"version_source"`
	Version         string `json:"version"`
	Cache           bool   `json:"cache"`
	AdapterRequired bool   `json:"adapter_required"`
}

type commandMapping struct {
	SourceCommand   string `json:"source_command"`
	NativeCommand   string `json:"native_command,omitempty"`
	NativeKind      string `json:"native_kind"`
	EvidencePath    string `json:"evidence_path,omitempty"`
	AdapterRequired bool   `json:"adapter_required"`
}

type childNativeMapping struct {
	Checkout            checkoutMapping `json:"checkout"`
	GoToolchain         goMapping       `json:"go_toolchain"`
	RepositoryReadme    commandMapping  `json:"repository_readme,omitempty"`
	ExecutableKnowledge commandMapping  `json:"executable_knowledge,omitempty"`
	DependencyAllowlist commandMapping  `json:"dependency_allowlist,omitempty"`
	ContextMap          commandMapping  `json:"context_map,omitempty"`
	EvidenceArtifact    artifactMapping `json:"evidence_artifact"`
}

type childParityClaim struct {
	AllSourceStepsMapped              bool `json:"all_source_steps_mapped"`
	RequiredAdapterCount              int  `json:"required_adapter_count"`
	RepositoryReadmeCommandExact      bool `json:"repository_readme_command_exact,omitempty"`
	ExecutableKnowledgeCommandExact   bool `json:"executable_knowledge_command_exact,omitempty"`
	DependencyAllowlistBehaviorExact  bool `json:"dependency_allowlist_behavior_exact,omitempty"`
	ContextMapCommandExact            bool `json:"context_map_command_exact,omitempty"`
	SecureEvidencePermissions         bool `json:"secure_evidence_permissions"`
	SourceWorkflowEdited              bool `json:"source_workflow_edited"`
	SourceWorkflowExecutedByThisSlice bool `json:"source_workflow_executed_by_this_slice"`
}

type artifactMapping struct {
	Source           string   `json:"source"`
	NativeKind       string   `json:"native_kind"`
	Paths            []string `json:"paths"`
	Redaction        string   `json:"redaction"`
	RunWhen          string   `json:"run_when"`
	IfNoFilesFound   string   `json:"if_no_files_found"`
	ContentAddressed bool     `json:"content_addressed"`
	AdapterRequired  bool     `json:"adapter_required"`
}
