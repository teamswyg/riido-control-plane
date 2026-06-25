package main

import (
	"strings"
	"testing"
)

func TestBuildDispatchPlanRejectsExpiredSourceCommands(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-26T00:00:00Z")
	_, err := buildDispatchPlan(repoRootForTest(t), refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
		Status:        "refresh_required",
		GeneratedAt:   "2026-06-24T00:00:00Z",
		ExpiresAt:     "2026-06-25T00:00:00Z",
		Commands: []selectedRefreshCommand{{
			LoopID:  "closed_loop_candidate",
			Kind:    "refresh_workflow",
			Command: "gh workflow run loop-registry.yml --ref main",
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expired source command rejection, got %v", err)
	}
}

func TestBuildDispatchPlanRequiresSourceFreshnessWindow(t *testing.T) {
	_, err := buildDispatchPlan(repoRootForTest(t), refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
		Status:        "refresh_required",
	})
	if err == nil || !strings.Contains(err.Error(), "generated_at and expires_at") {
		t.Fatalf("expected missing source freshness rejection, got %v", err)
	}
}
