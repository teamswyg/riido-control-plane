package main

const (
	goCIModuleDownloadCommand = "go mod download"
	goCIBaselineSourceCommand = "mkdir -p out && go test ./tools/gocibaseline -count=1 && go run ./tools/gocibaseline -check-doc -evidence-out out/go-ci-baseline-evidence.json"
	goCIBaselineNativeCommand = "umask 077 && " + goCIBaselineSourceCommand
	goCILintInstallCommand    = "go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2"
	goCILintInstallNative     = "mkdir -p bin && GOBIN=\"$PWD/bin\" " + goCILintInstallCommand
	goCILintCommand           = "golangci-lint run ./... --timeout=5m"
	goCILintNativeCommand     = "bin/golangci-lint run ./... --timeout=5m"
	goCICoverageSourceCommand = "mkdir -p out && go test ./... -coverprofile=out/go-test-cover.out -covermode=count && go tool cover -func=out/go-test-cover.out | tee out/go-test-coverage-summary.txt"
	goCICoverageNativeCommand = goCICoverageSourceCommand + " && chmod 0600 out/go-test-cover.out out/go-test-coverage-summary.txt"
)

func verifyGoCIWorkflow(repoRoot string, child boundedChild) bool {
	value := child.Baseline
	raw, err := readRootFile(repoRoot, value.Workflow)
	required := []string{
		"name: Go CI", "actions/checkout@v7", "actions/setup-go@v6", "go-version-file: go.mod",
		"cache: true", "run: go mod download", "go test ./tools/gocibaseline -count=1",
		"go run ./tools/gocibaseline", "-check-doc", "-evidence-out out/go-ci-baseline-evidence.json",
		goCILintInstallCommand, goCILintCommand, "-coverprofile=out/go-test-cover.out",
		"-covermode=count", "go tool cover -func=out/go-test-cover.out | tee out/go-test-coverage-summary.txt",
		"actions/upload-artifact@v7", "if: always()", "out/go-ci-baseline-evidence.json",
		"out/go-test-cover.out", "out/go-test-coverage-summary.txt", "if-no-files-found: error",
	}
	return err == nil && value.SourceRevision == "cc55c5326490b3d2dfbc19da1cfeb019e0c9b3d2" &&
		value.Workflow == ".github/workflows/go-ci.yml" &&
		value.WorkflowSHA256 == "82a08f1a850273ce756c0988fddbdfbcdd5590e059b442c172a89a7d41b731c4" &&
		digest(raw) == value.WorkflowSHA256 && value.WorkflowName == "Go CI" &&
		value.Job == "lint-test" && value.TrackedWorkflowCount == 56 && containsAll(string(raw), required)
}

func verifyGoCIMapping(child boundedChild) bool {
	mapping, claim := child.NativeMapping, child.ParityClaim
	return mapping.Checkout.Source == "actions/checkout@v7" && mapping.Checkout.NativeKind == "checkout" &&
		!mapping.Checkout.AdapterRequired && mapping.GoToolchain.Source == "actions/setup-go@v6" &&
		mapping.GoToolchain.NativeKind == "worker_toolchain" && mapping.GoToolchain.VersionSource == "go.mod" &&
		mapping.GoToolchain.Version == "1.26.5" && mapping.GoToolchain.Cache && !mapping.GoToolchain.AdapterRequired &&
		verifyCommand(mapping.ModuleDownload, goCIModuleDownloadCommand, goCIModuleDownloadCommand, "") &&
		verifyCommand(mapping.GoCIBaseline, goCIBaselineSourceCommand, goCIBaselineNativeCommand, "out/go-ci-baseline-evidence.json") &&
		verifyCommand(mapping.LintInstall, goCILintInstallCommand, goCILintInstallNative, "") &&
		verifyCommand(mapping.Lint, goCILintCommand, goCILintNativeCommand, "") &&
		verifyCommand(mapping.Coverage, goCICoverageSourceCommand, goCICoverageNativeCommand, "out/go-test-cover.out") &&
		claim.AllSourceStepsMapped && claim.RequiredAdapterCount == 0 &&
		claim.ModuleDownloadCommandExact && claim.GoCIBaselineCommandExact && claim.LintInstallCommandExact &&
		claim.LintCommandExact && claim.CoverageCommandExact && claim.CoverageSummaryBehaviorExact &&
		claim.SecureEvidencePermissions && !claim.SourceWorkflowEdited && !claim.SourceWorkflowExecutedByThisSlice
}

func verifyCommand(value commandMapping, source, native, evidencePath string) bool {
	return value.SourceCommand == source && value.NativeCommand == native && value.NativeKind == "shell" &&
		value.EvidencePath == evidencePath && !value.AdapterRequired
}
