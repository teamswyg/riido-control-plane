package main

const (
	preCommitBaselineSource = "mkdir -p out && go test ./tools/precommitbaseline -count=1 && go run ./tools/precommitbaseline -check-doc -evidence-out out/pre-commit-baseline-evidence.json"
	preCommitBaselineNative = "umask 077 && " + preCommitBaselineSource
)

func verifyPreCommitWorkflow(repoRoot string, child boundedChild) bool {
	value := child.Baseline
	raw, err := readRootFile(repoRoot, value.Workflow)
	required := []string{
		"name: pre-commit-baseline", "actions/checkout@v7", "actions/setup-go@v6",
		"go-version-file: go.mod", "cache: false", "go test ./tools/precommitbaseline -count=1",
		"go run ./tools/precommitbaseline", "-check-doc",
		"-evidence-out out/pre-commit-baseline-evidence.json", "actions/upload-artifact@v7",
		"if: always()", "name: pre-commit-baseline-evidence",
		"path: out/pre-commit-baseline-evidence.json", "if-no-files-found: error",
	}
	return err == nil && value.SourceRevision == "b32994b3524b110419b25600f45cdacd47772150" &&
		value.Workflow == ".github/workflows/pre-commit-baseline.yml" &&
		value.WorkflowSHA256 == "338db1065406d91b511c7fd8838f7b18964aabadbea8e2b4039cc9aa8e89c104" &&
		digest(raw) == value.WorkflowSHA256 && value.WorkflowName == "pre-commit-baseline" &&
		value.Job == "pre-commit-baseline" && value.TrackedWorkflowCount == 56 &&
		containsAll(string(raw), required)
}

func verifyPreCommitMapping(child boundedChild) bool {
	mapping, claim := child.NativeMapping, child.ParityClaim
	return mapping.Checkout.Source == "actions/checkout@v7" && mapping.Checkout.NativeKind == "checkout" &&
		!mapping.Checkout.AdapterRequired && mapping.GoToolchain.Source == "actions/setup-go@v6" &&
		mapping.GoToolchain.NativeKind == "worker_toolchain" && mapping.GoToolchain.VersionSource == "go.mod" &&
		mapping.GoToolchain.Version == "1.26.5" && !mapping.GoToolchain.Cache && !mapping.GoToolchain.AdapterRequired &&
		verifyCommand(mapping.PreCommitBaseline, preCommitBaselineSource, preCommitBaselineNative,
			"out/pre-commit-baseline-evidence.json") && claim.AllSourceStepsMapped &&
		claim.RequiredAdapterCount == 0 && claim.PreCommitBaselineCommandExact &&
		claim.SecureEvidencePermissions && !claim.SourceWorkflowEdited &&
		!claim.SourceWorkflowExecutedByThisSlice
}

func verifyPreCommitArtifact(child boundedChild) bool {
	value := child.NativeMapping.EvidenceArtifact
	return value.Source == "actions/upload-artifact@v7" && value.NativeKind == "artifact" &&
		len(value.Paths) == 1 && value.Paths[0] == "out/pre-commit-baseline-evidence.json" &&
		value.Redaction == "required" && value.RunWhen == "always" && value.IfNoFilesFound == "error" &&
		value.ContentAddressed && !value.AdapterRequired
}
