package main

import "slices"

const (
	evidenceGraphTestSource = "go test ./tools/evidencegraph -count=1"
	evidenceGraphSource     = "impact_args=(); if [ -n \"$RIIDO_EVIDENCE_GRAPH_IMPACT_BASE\" ]; then impact_args=(-impact-base \"$RIIDO_EVIDENCE_GRAPH_IMPACT_BASE\"); fi; go run ./tools/evidencegraph -check-doc -github-annotations \"${impact_args[@]}\" -evidence-out out/evidence-graph-evidence.json"
	evidenceGraphNative     = "umask 077 && go test ./tools/evidencegraph -count=1 && mkdir -p out && if [ -n \"${RIIDO_EVIDENCE_GRAPH_IMPACT_BASE:-}\" ]; then go run ./tools/evidencegraph -check-doc -github-annotations -impact-base \"$RIIDO_EVIDENCE_GRAPH_IMPACT_BASE\" -evidence-out out/evidence-graph-evidence.json; else go run ./tools/evidencegraph -check-doc -github-annotations -evidence-out out/evidence-graph-evidence.json; fi"
)

func verifyEvidenceGraphWorkflow(repoRoot string, child boundedChild) bool {
	value := child.Baseline
	raw, err := readRootFile(repoRoot, value.Workflow)
	required := []string{
		"name: Evidence Graph", "workflow_dispatch:", "schedule:", "cron: \"29 20 * * *\"",
		"pull_request:", "push:", "actions/checkout@v7", "fetch-depth: 0",
		"actions/setup-go@v6", "go-version-file: go.mod", "cache: true",
		evidenceGraphTestSource, "mkdir -p out", "RIIDO_EVIDENCE_GRAPH_IMPACT_BASE:",
		"impact_args=(-impact-base \"$RIIDO_EVIDENCE_GRAPH_IMPACT_BASE\")",
		"go run ./tools/evidencegraph", "-check-doc", "-github-annotations",
		"-evidence-out out/evidence-graph-evidence.json", "actions/upload-artifact@v7",
		"name: evidence-graph-evidence", "path: out/evidence-graph-evidence.json",
		"if-no-files-found: error",
	}
	return err == nil && value.SourceRevision == "cbcbbe9e9e612011836c622c300a3b6cafa0f638" &&
		value.Workflow == ".github/workflows/evidence-graph.yml" &&
		value.WorkflowSHA256 == "518ada12cd1754f1550ab008d1782192b30da2d2c3674427c20e3b63f1ba6710" &&
		digest(raw) == value.WorkflowSHA256 && value.WorkflowName == "Evidence Graph" &&
		value.Job == "evidence-graph" && value.TrackedWorkflowCount == 56 &&
		containsAll(string(raw), required)
}

func verifyEvidenceGraphMapping(child boundedChild) bool {
	mapping, claim := child.NativeMapping, child.ParityClaim
	return mapping.Checkout.Source == "actions/checkout@v7" && mapping.Checkout.NativeKind == "checkout" &&
		mapping.Checkout.FullHistory && !mapping.Checkout.AdapterRequired &&
		mapping.GoToolchain.Source == "actions/setup-go@v6" &&
		mapping.GoToolchain.NativeKind == "worker_toolchain" && mapping.GoToolchain.VersionSource == "go.mod" &&
		mapping.GoToolchain.Version == "1.26.5" && mapping.GoToolchain.Cache &&
		!mapping.GoToolchain.AdapterRequired &&
		verifyCommand(mapping.EvidenceGraphTest, evidenceGraphTestSource, evidenceGraphTestSource, "") &&
		verifyCommand(mapping.EvidenceGraph, evidenceGraphSource, evidenceGraphNative,
			"out/evidence-graph-evidence.json") &&
		claim.AllSourceStepsMapped && claim.RequiredAdapterCount == 0 &&
		claim.EvidenceGraphTestCommandExact && claim.EvidenceGraphImpactBaseExact &&
		claim.EvidenceGraphAnnotationsExact && claim.SecureEvidencePermissions &&
		!claim.SourceWorkflowEdited && !claim.SourceWorkflowExecutedByThisSlice
}

func verifyEvidenceGraphArtifact(child boundedChild) bool {
	value := child.NativeMapping.EvidenceArtifact
	want := []string{"out/evidence-graph-evidence.json"}
	return value.Source == "actions/upload-artifact@v7" && value.NativeKind == "artifact" &&
		slices.Equal(value.Paths, want) && value.Redaction == "required" &&
		value.RunWhen == "on_success" && value.IfNoFilesFound == "error" &&
		value.ContentAddressed && !value.AdapterRequired
}
