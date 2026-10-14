package main

const (
	moduleDependencySourceCommand = "go run ./tools/dependencyallowlist -contract dependency_allowlist.riido.json"
	moduleDependencyNativeCommand = "mkdir -p out && go run ./tools/dependencyallowlist -contract dependency_allowlist.riido.json -evidence-out out/dependency-allowlist-evidence.json"
	moduleDecompositionSource     = "mkdir -p out && go test ./tools/moduledecomposition -count=1 && go run ./tools/moduledecomposition -check-doc -evidence-out out/module-decomposition-evidence.json"
	moduleDecompositionNative     = "umask 077 && " + moduleDecompositionSource
	moduleKnowledgeSource         = "go run ./tools/knowledgecoverage -check-doc -evidence-out out/executable-knowledge-coverage.json"
	moduleKnowledgeNative         = "umask 077 && " + moduleKnowledgeSource
)

func verifyModuleDecompositionWorkflow(repoRoot string, child boundedChild) bool {
	value := child.Baseline
	raw, err := readRootFile(repoRoot, value.Workflow)
	required := []string{
		"name: module-decomposition", "actions/checkout@v7", "actions/setup-go@v6",
		"go-version-file: go.mod", "cache: false", moduleDependencySourceCommand,
		"go test ./tools/moduledecomposition -count=1", "go run ./tools/moduledecomposition",
		"-check-doc", "-evidence-out out/module-decomposition-evidence.json",
		"go run ./tools/knowledgecoverage", "-evidence-out out/executable-knowledge-coverage.json",
		"actions/upload-artifact@v7", "if: always()", "out/module-decomposition-evidence.json",
		"out/executable-knowledge-coverage.json", "if-no-files-found: error",
	}
	return err == nil && value.SourceRevision == "4f4f142f639eeb6162049bf7bb7b6266a184c137" &&
		value.Workflow == ".github/workflows/module-decomposition.yml" &&
		value.WorkflowSHA256 == "16077a7807afe5b64787ab7faf0321685f8797f5833fc87c85a9e1ae0c504480" &&
		digest(raw) == value.WorkflowSHA256 && value.WorkflowName == "module-decomposition" &&
		value.Job == "module-decomposition" && value.TrackedWorkflowCount == 56 && containsAll(string(raw), required)
}

func verifyModuleDecompositionMapping(child boundedChild) bool {
	mapping, claim := child.NativeMapping, child.ParityClaim
	return mapping.Checkout.Source == "actions/checkout@v7" && mapping.Checkout.NativeKind == "checkout" &&
		!mapping.Checkout.AdapterRequired && mapping.GoToolchain.Source == "actions/setup-go@v6" &&
		mapping.GoToolchain.NativeKind == "worker_toolchain" && mapping.GoToolchain.VersionSource == "go.mod" &&
		mapping.GoToolchain.Version == "1.26.5" && !mapping.GoToolchain.Cache && !mapping.GoToolchain.AdapterRequired &&
		verifyCommand(mapping.DependencyAllowlist, moduleDependencySourceCommand, moduleDependencyNativeCommand,
			"out/dependency-allowlist-evidence.json") &&
		verifyCommand(mapping.ModuleDecomposition, moduleDecompositionSource, moduleDecompositionNative,
			"out/module-decomposition-evidence.json") &&
		verifyCommand(mapping.ExecutableKnowledge, moduleKnowledgeSource, moduleKnowledgeNative,
			"out/executable-knowledge-coverage.json") &&
		claim.AllSourceStepsMapped && claim.RequiredAdapterCount == 0 &&
		claim.DependencyAllowlistBehaviorExact && claim.ModuleDecompositionCommandExact &&
		claim.ExecutableKnowledgeCommandExact && claim.SecureEvidencePermissions &&
		!claim.SourceWorkflowEdited && !claim.SourceWorkflowExecutedByThisSlice
}
