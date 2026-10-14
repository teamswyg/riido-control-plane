package main

import "slices"

const (
	executableKnowledgeSource = "mkdir -p out && go test ./tools/knowledgecoverage -count=1 && go run ./tools/knowledgecoverage -check-doc -evidence-out out/executable-knowledge-coverage.json"
	executableKnowledgeNative = "umask 077 && " + executableKnowledgeSource
)

func verifyExecutableKnowledgeWorkflow(repoRoot string, child boundedChild) bool {
	value := child.Baseline
	raw, err := readRootFile(repoRoot, value.Workflow)
	required := []string{
		"name: executable-knowledge-coverage", "actions/checkout@v7", "actions/setup-go@v6",
		"go-version-file: go.mod", "cache: false", "mkdir -p out",
		"go test ./tools/knowledgecoverage -count=1", "go run ./tools/knowledgecoverage",
		"-check-doc", "-evidence-out out/executable-knowledge-coverage.json",
		"actions/upload-artifact@v7", "if: always()", "name: executable-knowledge-coverage",
		"path: out/executable-knowledge-coverage.json", "if-no-files-found: error",
	}
	return err == nil && value.SourceRevision == "5c1251925749faf51c95a1f567a2f41d866ee31d" &&
		value.Workflow == ".github/workflows/executable-knowledge-coverage.yml" &&
		value.WorkflowSHA256 == "722c8b947b5738d827a4196cbc1716a804a19740277d2ced1fda1dd97497fd04" &&
		digest(raw) == value.WorkflowSHA256 && value.WorkflowName == "executable-knowledge-coverage" &&
		value.Job == "executable-knowledge-coverage" && value.TrackedWorkflowCount == 56 &&
		containsAll(string(raw), required)
}

func verifyExecutableKnowledgeMapping(child boundedChild) bool {
	mapping, claim := child.NativeMapping, child.ParityClaim
	return mapping.Checkout.Source == "actions/checkout@v7" && mapping.Checkout.NativeKind == "checkout" &&
		!mapping.Checkout.AdapterRequired && mapping.GoToolchain.Source == "actions/setup-go@v6" &&
		mapping.GoToolchain.NativeKind == "worker_toolchain" && mapping.GoToolchain.VersionSource == "go.mod" &&
		mapping.GoToolchain.Version == "1.26.5" && !mapping.GoToolchain.Cache &&
		!mapping.GoToolchain.AdapterRequired &&
		verifyCommand(mapping.ExecutableKnowledge, executableKnowledgeSource, executableKnowledgeNative,
			"out/executable-knowledge-coverage.json") &&
		claim.AllSourceStepsMapped && claim.RequiredAdapterCount == 0 &&
		claim.ExecutableKnowledgeCommandExact && claim.SecureEvidencePermissions &&
		!claim.SourceWorkflowEdited && !claim.SourceWorkflowExecutedByThisSlice
}

func verifyExecutableKnowledgeArtifact(child boundedChild) bool {
	value := child.NativeMapping.EvidenceArtifact
	want := []string{"out/executable-knowledge-coverage.json"}
	return value.Source == "actions/upload-artifact@v7" && value.NativeKind == "artifact" &&
		slices.Equal(value.Paths, want) && value.Redaction == "required" && value.RunWhen == "always" &&
		value.IfNoFilesFound == "error" && value.ContentAddressed && !value.AdapterRequired
}
