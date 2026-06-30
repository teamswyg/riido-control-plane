package main

import (
	"testing"
	"time"
)

func TestWorkflowEvidenceCarriesExpiry(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-24T00:00:00Z")

	result := auditResult{Records: []workflowRecord{{Path: ".github/workflows/example.yml"}}}
	got := newEvidence(manifest{ID: "workflow-evidence"}, result)

	if got.GeneratedAt != "2026-06-24T00:00:00Z" {
		t.Fatalf("generated_at = %q", got.GeneratedAt)
	}
	if got.ExpiresAt != "2026-06-25T00:00:00Z" {
		t.Fatalf("expires_at = %q", got.ExpiresAt)
	}
	if got.EvidenceTTLHours != 24 {
		t.Fatalf("evidence_ttl_hours = %d", got.EvidenceTTLHours)
	}
	if !got.WorkflowScheduled {
		t.Fatal("workflow_scheduled must be true")
	}
	if !got.StrictArtifactUpload {
		t.Fatal("strict_artifact_upload must be true")
	}
	generatedAt, err := time.Parse(time.RFC3339, got.GeneratedAt)
	if err != nil {
		t.Fatal(err)
	}
	expiresAt, err := time.Parse(time.RFC3339, got.ExpiresAt)
	if err != nil {
		t.Fatal(err)
	}
	if !generatedAt.Before(expiresAt) {
		t.Fatalf("expected generated_at before expires_at, got %s >= %s", got.GeneratedAt, got.ExpiresAt)
	}
}
