package main

import "strings"

func verifyLegacyWorkflow(repoRoot string, document manifest) bool {
	value := document.Baseline
	raw, err := readRootFile(repoRoot, value.Workflow)
	if err != nil {
		return false
	}
	required := []string{
		"actions/checkout@v7", "actions/setup-go@v6", "go-version-file: go.mod",
		"cache: false", "go run ./tools/dependencyallowlist",
		"-evidence-out out/dependency-allowlist-evidence.json",
		"actions/upload-artifact@v7", "if: always()", "if-no-files-found: error",
		"run: go test ./...",
	}
	return value.SourceRevision == "0191414e5dc7ac2f09c754dc8a731f57f8478371" &&
		value.Workflow == ".github/workflows/ci.yml" &&
		value.WorkflowSHA256 == "74430b5f9001bdd6d384c4bae985d7914f825f998b47ffe22826f480e4789f1c" &&
		digest(raw) == value.WorkflowSHA256 && value.WorkflowName == "ci" &&
		value.Job == "go" && value.TrackedWorkflowCount == 56 && containsAll(string(raw), required)
}

func verifyGoToolchain(repoRoot string, document manifest) bool {
	value := document.NativeMapping.GoToolchain
	raw, err := readRootFile(repoRoot, "go.mod")
	return err == nil && value.Source == "actions/setup-go@v6" &&
		value.NativeKind == "worker_toolchain" && value.VersionSource == "go.mod" &&
		value.Version == "1.26.5" && !value.Cache && !value.AdapterRequired &&
		strings.Contains("\n"+string(raw), "\ngo 1.26.5\n")
}
