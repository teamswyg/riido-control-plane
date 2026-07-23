package main

import "slices"

const (
	integrationMatrixDependencySource = "go run ./tools/dependencyallowlist -contract dependency_allowlist.riido.json"
	integrationMatrixDependencyNative = "mkdir -p out && go run ./tools/dependencyallowlist -contract dependency_allowlist.riido.json -evidence-out out/dependency-allowlist-evidence.json"
	integrationMatrixSource           = "mkdir -p out && go test ./tools/integrationmatrix -count=1 && go run ./tools/integrationmatrix -check-doc -evidence-out out/integration-matrix-evidence.json"
	integrationMatrixNative           = "umask 077 && " + integrationMatrixSource
	integrationMatrixKnowledgeSource  = "go run ./tools/knowledgecoverage -check-doc -evidence-out out/executable-knowledge-coverage.json"
	integrationMatrixKnowledgeNative  = "umask 077 && " + integrationMatrixKnowledgeSource
)

func verifyIntegrationMatrixWorkflow(repoRoot string, child boundedChild) bool {
	value := child.Baseline
	raw, err := readRootFile(repoRoot, value.Workflow)
	required := []string{
		"name: integration-matrix", "actions/checkout@v7", "actions/setup-go@v6",
		"go-version-file: go.mod", "cache: false", integrationMatrixDependencySource,
		"go test ./tools/integrationmatrix -count=1", "go run ./tools/integrationmatrix",
		"-check-doc", "-evidence-out out/integration-matrix-evidence.json",
		"go run ./tools/knowledgecoverage", "-evidence-out out/executable-knowledge-coverage.json",
		"actions/upload-artifact@v7", "if: always()", "name: integration-matrix-evidence",
		"out/integration-matrix-evidence.json", "out/executable-knowledge-coverage.json",
		"if-no-files-found: error",
	}
	return err == nil && value.SourceRevision == "7cb9611494891674976d584d52f927f162c5f82e" &&
		value.Workflow == ".github/workflows/integration-matrix.yml" &&
		value.WorkflowSHA256 == "0d2df869bc0836ba70ac9d109d14c03ef9b927ac960d85a699190fc5d6ed9f61" &&
		digest(raw) == value.WorkflowSHA256 && value.WorkflowName == "integration-matrix" &&
		value.Job == "integration-matrix" && value.TrackedWorkflowCount == 56 &&
		containsAll(string(raw), required)
}

func verifyIntegrationMatrixMapping(child boundedChild) bool {
	mapping, claim := child.NativeMapping, child.ParityClaim
	return mapping.Checkout.Source == "actions/checkout@v7" && mapping.Checkout.NativeKind == "checkout" &&
		!mapping.Checkout.AdapterRequired && mapping.GoToolchain.Source == "actions/setup-go@v6" &&
		mapping.GoToolchain.NativeKind == "worker_toolchain" && mapping.GoToolchain.VersionSource == "go.mod" &&
		mapping.GoToolchain.Version == "1.26.5" && !mapping.GoToolchain.Cache &&
		!mapping.GoToolchain.AdapterRequired &&
		verifyCommand(mapping.DependencyAllowlist, integrationMatrixDependencySource,
			integrationMatrixDependencyNative, "out/dependency-allowlist-evidence.json") &&
		verifyCommand(mapping.IntegrationMatrix, integrationMatrixSource, integrationMatrixNative,
			"out/integration-matrix-evidence.json") &&
		verifyCommand(mapping.ExecutableKnowledge, integrationMatrixKnowledgeSource,
			integrationMatrixKnowledgeNative, "out/executable-knowledge-coverage.json") &&
		claim.AllSourceStepsMapped && claim.RequiredAdapterCount == 0 &&
		claim.DependencyAllowlistBehaviorExact && claim.IntegrationMatrixCommandExact &&
		claim.ExecutableKnowledgeCommandExact && claim.SecureEvidencePermissions &&
		!claim.SourceWorkflowEdited && !claim.SourceWorkflowExecutedByThisSlice
}

func verifyIntegrationMatrixArtifact(child boundedChild) bool {
	value := child.NativeMapping.EvidenceArtifact
	want := []string{"out/integration-matrix-evidence.json", "out/executable-knowledge-coverage.json"}
	return value.Source == "actions/upload-artifact@v7" && value.NativeKind == "artifact" &&
		slices.Equal(value.Paths, want) && value.Redaction == "required" && value.RunWhen == "always" &&
		value.IfNoFilesFound == "error" && value.ContentAddressed && !value.AdapterRequired
}
