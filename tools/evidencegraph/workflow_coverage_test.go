package main

import "testing"

func TestEvidenceWorkflowCoversReferencedPaths(t *testing.T) {
	m, err := loadManifest("../../" + defaultManifest)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := verifyEvidenceWorkflowCoversRefs("../..", m); err != nil {
		t.Fatalf("workflow coverage: %v", err)
	}
}

func TestEvidenceWorkflowCoverageFailsForMissingPath(t *testing.T) {
	m, err := loadManifest("../../" + defaultManifest)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	m.Chains[0].Changes = append(m.Chains[0].Changes, ref{Kind: "code", Path: "internal/missing.go"})
	if err := verifyEvidenceWorkflowCoversRefs("../..", m); err == nil {
		t.Fatal("expected missing workflow path to fail")
	}
}

func TestWorkflowPathFiltersExtractsAllPathBlocks(t *testing.T) {
	got := workflowPathFilters("push:\n  paths:\n    - \"a.go\"\npull_request:\n  paths:\n    - 'b.go'\n")
	if !got["a.go"] || !got["b.go"] {
		t.Fatalf("paths = %+v", got)
	}
}

func TestWorkflowPathCoveredSupportsDirectoryGlob(t *testing.T) {
	filters := map[string]bool{"tools/evidencegraph/**": true}
	if !workflowPathCovered(filters, "tools/evidencegraph") {
		t.Fatal("expected directory glob to cover package path")
	}
}
