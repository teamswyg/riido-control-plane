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
