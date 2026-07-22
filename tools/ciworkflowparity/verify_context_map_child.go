package main

import (
	"slices"
)

const (
	contextDependencySourceCommand = "go run ./tools/dependencyallowlist -contract dependency_allowlist.riido.json"
	contextDependencyNativeCommand = "mkdir -p out && go run ./tools/dependencyallowlist -contract dependency_allowlist.riido.json -evidence-out out/dependency-allowlist-evidence.json"
	contextMapSourceCommand        = "mkdir -p out && go test ./tools/contextmap -count=1 && go run ./tools/contextmap -check-doc -evidence-out out/control-plane-context-map-evidence.json"
	contextMapNativeCommand        = "umask 077 && " + contextMapSourceCommand
)

func verifyContextMapWorkflow(repoRoot string, child boundedChild) bool {
	value := child.Baseline
	raw, err := readRootFile(repoRoot, value.Workflow)
	required := []string{
		"name: context-map", "actions/checkout@v7", "actions/setup-go@v6",
		"go-version-file: go.mod", "cache: false", contextDependencySourceCommand,
		"go test ./tools/contextmap -count=1", "go run ./tools/contextmap", "-check-doc",
		"-evidence-out out/control-plane-context-map-evidence.json", "actions/upload-artifact@v7",
		"if: always()", "if-no-files-found: error",
	}
	return err == nil && value.SourceRevision == "86930893218cf31124b9f177a30442cb52bb3fce" &&
		value.Workflow == ".github/workflows/context-map.yml" &&
		value.WorkflowSHA256 == "3abfc4a36a0e9121996116a8010b56769a75d30915c64d394761c552ccae5392" &&
		digest(raw) == value.WorkflowSHA256 && value.WorkflowName == "context-map" &&
		value.Job == "context-map" && value.TrackedWorkflowCount == 56 && containsAll(string(raw), required)
}

func verifyContextMapMapping(child boundedChild) bool {
	mapping, claim := child.NativeMapping, child.ParityClaim
	return mapping.Checkout.Source == "actions/checkout@v7" && mapping.Checkout.NativeKind == "checkout" &&
		!mapping.Checkout.AdapterRequired && mapping.GoToolchain.Source == "actions/setup-go@v6" &&
		mapping.GoToolchain.NativeKind == "worker_toolchain" && mapping.GoToolchain.VersionSource == "go.mod" &&
		mapping.GoToolchain.Version == "1.26.5" && !mapping.GoToolchain.Cache && !mapping.GoToolchain.AdapterRequired &&
		mapping.DependencyAllowlist.SourceCommand == contextDependencySourceCommand &&
		mapping.DependencyAllowlist.NativeCommand == contextDependencyNativeCommand &&
		mapping.DependencyAllowlist.NativeKind == "shell" &&
		mapping.DependencyAllowlist.EvidencePath == "out/dependency-allowlist-evidence.json" &&
		!mapping.DependencyAllowlist.AdapterRequired && mapping.ContextMap.SourceCommand == contextMapSourceCommand &&
		mapping.ContextMap.NativeCommand == contextMapNativeCommand && mapping.ContextMap.NativeKind == "shell" &&
		mapping.ContextMap.EvidencePath == "out/control-plane-context-map-evidence.json" &&
		!mapping.ContextMap.AdapterRequired && claim.AllSourceStepsMapped && claim.RequiredAdapterCount == 0 &&
		claim.DependencyAllowlistBehaviorExact && claim.ContextMapCommandExact && claim.SecureEvidencePermissions &&
		!claim.SourceWorkflowEdited && !claim.SourceWorkflowExecutedByThisSlice
}

func verifyContextMapArtifact(child boundedChild) bool {
	value := child.NativeMapping.EvidenceArtifact
	want := []string{"out/control-plane-context-map-evidence.json"}
	return value.Source == "actions/upload-artifact@v7" && value.NativeKind == "artifact" &&
		slices.Equal(value.Paths, want) && value.Redaction == "required" && value.RunWhen == "always" &&
		value.IfNoFilesFound == "error" && value.ContentAddressed && !value.AdapterRequired
}

func verifyContextMapPipeline(repoRoot string, document manifest, child boundedChild) bool {
	var value pipeline
	if err := readJSON(repoRoot, document.Runner.Pipeline, &value); err != nil || len(value.Steps) != pipelineSteps {
		return false
	}
	steps := value.Steps
	return steps[2].ID == "verify-go-dependency-allowlist" && steps[2].Kind == "shell" &&
		steps[2].Command == child.NativeMapping.DependencyAllowlist.NativeCommand &&
		steps[8].ID == "verify-context-map" && steps[8].Kind == "shell" &&
		steps[8].Command == child.NativeMapping.ContextMap.NativeCommand &&
		steps[9].ID == "collect-context-map-evidence" && steps[9].Kind == "artifact" &&
		slices.Equal(steps[9].Paths, child.NativeMapping.EvidenceArtifact.Paths) &&
		steps[9].Redaction == "required" && steps[9].RunWhen == "always" &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/contextmap")
}
