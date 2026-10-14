package main

import "slices"

const (
	loopClosureAuditLoopID = "loop_closure_audit"
	loopClosureAuditSource = "mkdir -p out && go test ./tools/loopclosureaudit -count=1 && go run ./tools/loopclosureaudit -check-doc -evidence-out out/loop-closure-audit-evidence.json -candidate-out out/loop-closure-audit-closed-loop-candidates.json"
	loopClosureAuditNative = "umask 077 && export RIIDO_LOOP_IDS=" + loopClosureAuditLoopID + " && " + loopClosureAuditSource
)

func verifyLoopClosureAuditWorkflow(repoRoot string, child boundedChild) bool {
	value := child.Baseline
	raw, err := readRootFile(repoRoot, value.Workflow)
	required := []string{
		"name: loop-closure-audit", "workflow_dispatch:", "schedule:", "cron: \"11 21 * * *\"",
		"RIIDO_LOOP_IDS: loop_closure_audit", "actions/checkout@v7", "actions/setup-go@v6",
		"go-version-file: go.mod", "cache: false", "mkdir -p out",
		"go test ./tools/loopclosureaudit -count=1", "go run ./tools/loopclosureaudit",
		"-check-doc", "-evidence-out out/loop-closure-audit-evidence.json",
		"-candidate-out out/loop-closure-audit-closed-loop-candidates.json",
		"actions/upload-artifact@v7", "if: always()", "name: loop-closure-audit-evidence",
		"path: out/loop-closure-audit-evidence.json",
		"name: loop-closure-audit-closed-loop-candidates",
		"path: out/loop-closure-audit-closed-loop-candidates.json", "if-no-files-found: error",
	}
	return err == nil && value.SourceRevision == "feffdbf7faa12c6d343173a83989aacf56159f85" &&
		value.Workflow == ".github/workflows/loop-closure-audit.yml" &&
		value.WorkflowSHA256 == "a26fbcca723421a900bdddd2cba8da71079d1110c77acd561fe2e626ae601781" &&
		digest(raw) == value.WorkflowSHA256 && value.WorkflowName == "loop-closure-audit" &&
		value.Job == "audit" && value.TrackedWorkflowCount == 56 &&
		containsAll(string(raw), required)
}

func verifyLoopClosureAuditMapping(child boundedChild) bool {
	mapping, claim := child.NativeMapping, child.ParityClaim
	return mapping.Checkout.Source == "actions/checkout@v7" && mapping.Checkout.NativeKind == "checkout" &&
		!mapping.Checkout.AdapterRequired && mapping.GoToolchain.Source == "actions/setup-go@v6" &&
		mapping.GoToolchain.NativeKind == "worker_toolchain" && mapping.GoToolchain.VersionSource == "go.mod" &&
		mapping.GoToolchain.Version == "1.26.5" && !mapping.GoToolchain.Cache &&
		!mapping.GoToolchain.AdapterRequired &&
		verifyCommand(mapping.LoopClosureAudit, loopClosureAuditSource, loopClosureAuditNative,
			"out/loop-closure-audit-evidence.json") &&
		claim.AllSourceStepsMapped && claim.RequiredAdapterCount == 0 &&
		claim.LoopClosureAuditCommandExact && claim.LoopClosureAuditLoopEnvironmentExact &&
		claim.SecureEvidencePermissions && !claim.SourceWorkflowEdited &&
		!claim.SourceWorkflowExecutedByThisSlice
}

func verifyLoopClosureAuditArtifact(child boundedChild) bool {
	value := child.NativeMapping.EvidenceArtifact
	want := []string{
		"out/loop-closure-audit-evidence.json",
		"out/loop-closure-audit-closed-loop-candidates.json",
	}
	return value.Source == "actions/upload-artifact@v7" && value.NativeKind == "artifact" &&
		slices.Equal(value.Paths, want) && value.Redaction == "required" && value.RunWhen == "always" &&
		value.IfNoFilesFound == "error" && value.ContentAddressed && !value.AdapterRequired
}
