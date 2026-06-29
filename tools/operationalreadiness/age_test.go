package main

import (
	"testing"
	"time"
)

func TestOperationalReadinessEvidenceTracksPartialAge(t *testing.T) {
	m := manifest{SchemaVersion: manifestSchema, RequiredCategories: []string{"stress"}, Checks: []readinessCheck{
		{ID: "old", Date: "2026-06-26", Category: "stress", Status: "partial"},
		{ID: "fresh", Date: "2026-06-29", Category: "stress", Status: "partial"},
		{ID: "done", Date: "2026-06-26", Category: "stress", Status: "covered"},
	}}
	got := newEvidenceAt(m, time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC))
	if got.StalePartialCount != 1 || got.StaleAfterDays != stalePartialAfterDays {
		t.Fatalf("stale partial evidence = %+v", got)
	}
	if got.PartialChecks[0].AgeDays != 3 || !got.PartialChecks[0].Stale {
		t.Fatalf("old partial evidence = %+v", got.PartialChecks[0])
	}
	if got.PartialChecks[1].AgeDays != 0 || got.PartialChecks[1].Stale {
		t.Fatalf("fresh partial evidence = %+v", got.PartialChecks[1])
	}
}
