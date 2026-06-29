package main

import (
	"testing"
	"time"
)

func TestOperationalReadinessCandidateAdoptsPartialNextWorkFirst(t *testing.T) {
	m := candidateManifestWithPartial("old", "old_evidence", "go test ./old", "2026-06-26")
	now := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	item := newCandidateEvidence(m, newEvidenceAt(m, now), now).Candidates[0]
	if item.RequiredNextArtifacts[0] != "old_evidence" {
		t.Fatalf("first required artifact = %q", item.RequiredNextArtifacts[0])
	}
	first := item.AdoptionPlan[0]
	if first.Artifact != "old_evidence" || first.Command != "go test ./old" {
		t.Fatalf("first adoption step = %+v", first)
	}
	if len(item.AdoptionPlan) != len(candidateRequiredArtifacts())+1 {
		t.Fatalf("adoption plan = %+v", item.AdoptionPlan)
	}
}

func candidateManifestWithPartial(id, artifact, command, date string) manifest {
	return manifest{
		SchemaVersion: manifestSchema,
		Workflow:      ".github/workflows/operational-readiness.yml",
		Sources:       []producerSource{testCandidateSource()},
		Checks: []readinessCheck{{
			ID:           id,
			Date:         date,
			Category:     "stress",
			Status:       "partial",
			NextArtifact: artifact,
			NextCommand:  command,
		}},
	}
}
