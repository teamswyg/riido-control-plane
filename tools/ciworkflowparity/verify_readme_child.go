package main

import "strings"

const (
	readmeSourceCommand    = "mkdir -p out && go test ./tools/repositoryreadme -count=1 && go run ./tools/repositoryreadme -check-doc -evidence-out out/repository-readme-evidence.json"
	readmeNativeCommand    = "umask 077 && " + readmeSourceCommand
	knowledgeSourceCommand = "go run ./tools/knowledgecoverage -check-doc -evidence-out out/executable-knowledge-coverage.json"
	knowledgeNativeCommand = "umask 077 && " + knowledgeSourceCommand
)

func verifyReadmeWorkflow(repoRoot string, child boundedChild) bool {
	value := child.Baseline
	raw, err := readRootFile(repoRoot, value.Workflow)
	required := []string{
		"name: repository-readme", "actions/checkout@v7", "actions/setup-go@v6",
		"go-version-file: go.mod", "cache: false", "go test ./tools/repositoryreadme -count=1",
		"go run ./tools/repositoryreadme", "-check-doc", "-evidence-out out/repository-readme-evidence.json",
		"go run ./tools/knowledgecoverage", "-evidence-out out/executable-knowledge-coverage.json",
		"actions/upload-artifact@v7", "if: always()", "if-no-files-found: error",
	}
	return err == nil && value.SourceRevision == "646abfb4707d2addc47d37fd8a384e343c64a28c" &&
		value.Workflow == ".github/workflows/repository-readme.yml" &&
		value.WorkflowSHA256 == "ae953f656aba408e154b10a7d4bd782169fa51fc93ddd618e2d3313a5b82e58c" &&
		digest(raw) == value.WorkflowSHA256 && value.WorkflowName == "repository-readme" &&
		value.Job == "repository-readme" && value.TrackedWorkflowCount == 56 && containsAll(string(raw), required)
}

func verifyReadmeMapping(child boundedChild) bool {
	mapping, claim := child.NativeMapping, child.ParityClaim
	return mapping.Checkout.Source == "actions/checkout@v7" && mapping.Checkout.NativeKind == "checkout" &&
		!mapping.Checkout.AdapterRequired && mapping.GoToolchain.Source == "actions/setup-go@v6" &&
		mapping.GoToolchain.NativeKind == "worker_toolchain" && mapping.GoToolchain.VersionSource == "go.mod" &&
		mapping.GoToolchain.Version == "1.26.5" && !mapping.GoToolchain.Cache && !mapping.GoToolchain.AdapterRequired &&
		mapping.RepositoryReadme.SourceCommand == readmeSourceCommand &&
		mapping.RepositoryReadme.NativeCommand == readmeNativeCommand &&
		mapping.RepositoryReadme.NativeKind == "shell" &&
		mapping.RepositoryReadme.EvidencePath == "out/repository-readme-evidence.json" &&
		!mapping.RepositoryReadme.AdapterRequired &&
		mapping.ExecutableKnowledge.SourceCommand == knowledgeSourceCommand &&
		mapping.ExecutableKnowledge.NativeCommand == knowledgeNativeCommand &&
		mapping.ExecutableKnowledge.NativeKind == "shell" &&
		mapping.ExecutableKnowledge.EvidencePath == "out/executable-knowledge-coverage.json" &&
		!mapping.ExecutableKnowledge.AdapterRequired && claim.AllSourceStepsMapped &&
		claim.RequiredAdapterCount == 0 && claim.RepositoryReadmeCommandExact &&
		claim.ExecutableKnowledgeCommandExact && claim.SecureEvidencePermissions &&
		!claim.SourceWorkflowEdited && !claim.SourceWorkflowExecutedByThisSlice
}

func verifyAuthorityForChild(child boundedChild) bool {
	return verifyAuthority(manifest{Authority: child.Authority}) &&
		child.Rollback.BaselineWorkflowPreserved &&
		child.Rollback.Method == "retain_exact_baseline_until_separate_retirement_parity_review" &&
		child.Rollback.BlockedToPassedPrivatePolicyPairRequired
}

func verifyReadmeClassification(child boundedChild) bool {
	return child.Classification.Code == "repository_readme_native_parity_source_complete" &&
		strings.Contains(child.Classification.Meaning, "repository README") &&
		len(child.Classification.DoesNotClaim) == 5
}
