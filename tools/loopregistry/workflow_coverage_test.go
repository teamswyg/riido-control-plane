package main

import "testing"

func TestRegistryWorkflowCoversClaimBoundPaths(t *testing.T) {
	m, _ := loadLoopRegistryForTest(t)
	if err := verifyRegistryWorkflowCoversClaims("../..", m); err != nil {
		t.Fatalf("workflow coverage: %v", err)
	}
}

func TestRegistryWorkflowCoverageFailsForMissingClaimPath(t *testing.T) {
	m, _ := loadLoopRegistryForTest(t)
	m.Claims[0].Files = append(m.Claims[0].Files, "internal/riidoaiserver/missing_bound_surface.go")
	if err := verifyRegistryWorkflowCoversClaims("../..", m); err == nil {
		t.Fatal("expected missing claim-bound workflow path to fail")
	}
}

func TestWorkflowPathFiltersExtractsAllPathBlocks(t *testing.T) {
	got := workflowPathFilters("push:\n  paths:\n    - \"a.go\"\npull_request:\n  paths:\n    - 'b.go'\n")
	if !got["a.go"] || !got["b.go"] {
		t.Fatalf("paths = %+v", got)
	}
}

func TestWorkflowPathCoveredSupportsDirectoryGlob(t *testing.T) {
	filters := map[string]bool{"tools/loopregistry/**": true}
	if !workflowPathCovered(filters, "tools/loopregistry/workflow_coverage.go") {
		t.Fatal("expected directory glob to cover nested file")
	}
}
