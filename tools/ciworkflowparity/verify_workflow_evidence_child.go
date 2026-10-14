package main

import "slices"

const (
	workflowEvidenceLoopID = "closed_loop_candidate"
	workflowEvidenceSource = "mkdir -p out && go test ./tools/workflowevidence -count=1 && go run ./tools/workflowevidence -check-doc -evidence-out out/workflow-evidence.json"
	workflowEvidenceNative = "umask 077 && export RIIDO_LOOP_IDS=" + workflowEvidenceLoopID + " && " + workflowEvidenceSource
)

func verifyWorkflowEvidenceWorkflow(repoRoot string, child boundedChild) bool {
	value := child.Baseline
	raw, err := readRootFile(repoRoot, value.Workflow)
	required := []string{
		"name: workflow-evidence", "RIIDO_LOOP_IDS: closed_loop_candidate",
		"actions/checkout@v7", "actions/setup-go@v6", "go-version-file: go.mod",
		"cache: false", "mkdir -p out", "go test ./tools/workflowevidence -count=1",
		"go run ./tools/workflowevidence", "-check-doc",
		"-evidence-out out/workflow-evidence.json", "actions/upload-artifact@v7",
		"if: always()", "name: workflow-evidence", "path: out/workflow-evidence.json",
		"if-no-files-found: error",
	}
	return err == nil && value.SourceRevision == "632f2edd52b4aaf02bda141351221b5d0dac3e6a" &&
		value.Workflow == ".github/workflows/workflow-evidence.yml" &&
		value.WorkflowSHA256 == "daf9a9462e10401ae538e151f0d54daec59438cffc0ff2fce935e3834c28028b" &&
		digest(raw) == value.WorkflowSHA256 && value.WorkflowName == "workflow-evidence" &&
		value.Job == "workflow-evidence" && value.TrackedWorkflowCount == 56 &&
		containsAll(string(raw), required)
}

func verifyWorkflowEvidenceMapping(child boundedChild) bool {
	mapping, claim := child.NativeMapping, child.ParityClaim
	return mapping.Checkout.Source == "actions/checkout@v7" && mapping.Checkout.NativeKind == "checkout" &&
		!mapping.Checkout.AdapterRequired && mapping.GoToolchain.Source == "actions/setup-go@v6" &&
		mapping.GoToolchain.NativeKind == "worker_toolchain" && mapping.GoToolchain.VersionSource == "go.mod" &&
		mapping.GoToolchain.Version == "1.26.5" && !mapping.GoToolchain.Cache &&
		!mapping.GoToolchain.AdapterRequired &&
		verifyCommand(mapping.WorkflowEvidence, workflowEvidenceSource, workflowEvidenceNative,
			"out/workflow-evidence.json") &&
		claim.AllSourceStepsMapped && claim.RequiredAdapterCount == 0 &&
		claim.WorkflowEvidenceCommandExact && claim.WorkflowEvidenceLoopEnvironmentExact &&
		claim.SecureEvidencePermissions && !claim.SourceWorkflowEdited &&
		!claim.SourceWorkflowExecutedByThisSlice
}

func verifyWorkflowEvidenceArtifact(child boundedChild) bool {
	value := child.NativeMapping.EvidenceArtifact
	want := []string{"out/workflow-evidence.json"}
	return value.Source == "actions/upload-artifact@v7" && value.NativeKind == "artifact" &&
		slices.Equal(value.Paths, want) && value.Redaction == "required" && value.RunWhen == "always" &&
		value.IfNoFilesFound == "error" && value.ContentAddressed && !value.AdapterRequired
}
