package main

import (
	"testing"
	"time"
)

func TestOperationalReadinessCandidatesUseOnlyStalePartials(t *testing.T) {
	m := manifest{
		SchemaVersion:    manifestSchema,
		Workflow:         ".github/workflows/operational-readiness.yml",
		EvidenceArtifact: "custom-readiness-evidence",
		Sources:          []producerSource{testCandidateSource()},
		Checks: []readinessCheck{
			{
				ID:           "old",
				Date:         "2026-06-26",
				Category:     "stress",
				Status:       "partial",
				NextArtifact: "old_evidence",
				NextCommand:  "go test ./old",
			},
			{
				ID:           "fresh",
				Date:         "2026-06-29",
				Category:     "stress",
				Status:       "partial",
				NextArtifact: "fresh_evidence",
				NextCommand:  "go test ./fresh",
			},
		},
	}
	now := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	got := newCandidateEvidence(m, newEvidenceAt(m, now), now)
	if got.CandidateCount != 1 || got.Candidates[0].ID != "operational-readiness:old" {
		t.Fatalf("candidate evidence = %+v", got)
	}
	item := got.Candidates[0]
	if item.SourceRef.SummaryArtifact != m.EvidenceArtifact ||
		item.SourceRef.CandidateArtifact != readinessCandidateArtifact ||
		item.PromotionEdge.To != readinessPromotionTarget ||
		len(item.AdoptionPlan) != len(candidateRequiredArtifacts())+1 {
		t.Fatalf("candidate item = %+v", item)
	}
}
