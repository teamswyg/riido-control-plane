package main

import (
	"testing"
	"time"
)

func TestOperationalReadinessEvidenceExposesPartialPromotion(t *testing.T) {
	m := manifest{Checks: []readinessCheck{
		{ID: "old", Date: "2026-06-26", Category: "stress", Status: "partial"},
		{ID: "fresh", Date: "2026-06-29", Category: "stress", Status: "partial"},
	}}
	now := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	got := newEvidenceAt(m, now).PartialPromotion
	if got.CandidateArtifact != readinessCandidateArtifact ||
		got.CandidateCount != 1 ||
		got.CandidateIDs[0] != "operational-readiness:old" ||
		got.StalePartialIDs[0] != "old" {
		t.Fatalf("partial promotion = %+v", got)
	}
}
