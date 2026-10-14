package main

import "slices"

const (
	configReferenceDependencySource = "go run ./tools/dependencyallowlist -contract dependency_allowlist.riido.json"
	configReferenceDependencyNative = "mkdir -p out && go run ./tools/dependencyallowlist -contract dependency_allowlist.riido.json -evidence-out out/dependency-allowlist-evidence.json"
	configReferenceSource           = "mkdir -p out && go test ./tools/configreference -count=1 && go run ./tools/configreference -check-doc -evidence-out out/config-reference-evidence.json"
	configReferenceNative           = "umask 077 && " + configReferenceSource
	configReferenceKnowledgeSource  = "go run ./tools/knowledgecoverage -check-doc -evidence-out out/executable-knowledge-coverage.json"
	configReferenceKnowledgeNative  = "umask 077 && " + configReferenceKnowledgeSource
)

func verifyConfigReferenceWorkflow(repoRoot string, child boundedChild) bool {
	value := child.Baseline
	raw, err := readRootFile(repoRoot, value.Workflow)
	required := []string{
		"name: config-reference", "actions/checkout@v7", "actions/setup-go@v6",
		"go-version-file: go.mod", "cache: false", configReferenceDependencySource,
		"go test ./tools/configreference -count=1", "go run ./tools/configreference",
		"-check-doc", "-evidence-out out/config-reference-evidence.json",
		"go run ./tools/knowledgecoverage", "-evidence-out out/executable-knowledge-coverage.json",
		"actions/upload-artifact@v7", "if: always()", "name: config-reference-evidence",
		"out/config-reference-evidence.json", "out/executable-knowledge-coverage.json",
		"if-no-files-found: error",
	}
	return err == nil && value.SourceRevision == "d5033121528ac5ce5aa75eb33d287392b45baf6d" &&
		value.Workflow == ".github/workflows/config-reference.yml" &&
		value.WorkflowSHA256 == "5cbf30d7a80482fc5f29daface720c3418d31448aaac7d18a2b57c3744683d89" &&
		digest(raw) == value.WorkflowSHA256 && value.WorkflowName == "config-reference" &&
		value.Job == "config-reference" && value.TrackedWorkflowCount == 56 && containsAll(string(raw), required)
}

func verifyConfigReferenceMapping(child boundedChild) bool {
	mapping, claim := child.NativeMapping, child.ParityClaim
	return mapping.Checkout.Source == "actions/checkout@v7" && mapping.Checkout.NativeKind == "checkout" &&
		!mapping.Checkout.AdapterRequired && mapping.GoToolchain.Source == "actions/setup-go@v6" &&
		mapping.GoToolchain.NativeKind == "worker_toolchain" && mapping.GoToolchain.VersionSource == "go.mod" &&
		mapping.GoToolchain.Version == "1.26.5" && !mapping.GoToolchain.Cache && !mapping.GoToolchain.AdapterRequired &&
		verifyCommand(mapping.DependencyAllowlist, configReferenceDependencySource, configReferenceDependencyNative,
			"out/dependency-allowlist-evidence.json") &&
		verifyCommand(mapping.ConfigReference, configReferenceSource, configReferenceNative,
			"out/config-reference-evidence.json") &&
		verifyCommand(mapping.ExecutableKnowledge, configReferenceKnowledgeSource, configReferenceKnowledgeNative,
			"out/executable-knowledge-coverage.json") &&
		claim.AllSourceStepsMapped && claim.RequiredAdapterCount == 0 &&
		claim.DependencyAllowlistBehaviorExact && claim.ConfigReferenceCommandExact &&
		claim.ExecutableKnowledgeCommandExact && claim.SecureEvidencePermissions &&
		!claim.SourceWorkflowEdited && !claim.SourceWorkflowExecutedByThisSlice
}

func verifyConfigReferenceArtifact(child boundedChild) bool {
	value := child.NativeMapping.EvidenceArtifact
	want := []string{"out/config-reference-evidence.json", "out/executable-knowledge-coverage.json"}
	return value.Source == "actions/upload-artifact@v7" && value.NativeKind == "artifact" &&
		slices.Equal(value.Paths, want) && value.Redaction == "required" && value.RunWhen == "always" &&
		value.IfNoFilesFound == "error" && value.ContentAddressed && !value.AdapterRequired
}
