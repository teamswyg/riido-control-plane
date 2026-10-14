package main

import (
	"testing"
	"time"
)

func TestOperationalReadinessBackfillsNotionP0Cycles(t *testing.T) {
	m := loadManifestForTest(t)
	if err := verifyNotionOpenLoop("../..", m); err != nil {
		t.Fatal(err)
	}
	now, err := time.Parse(time.RFC3339, "2026-06-30T12:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	e := newEvidenceAt(m, now)
	if e.NotionOpenLoop.CycleCount < 6 || e.NotionOpenLoop.P0Count != e.NotionOpenLoop.CycleCount {
		t.Fatalf("notion p0 cycles missing: %+v", e.NotionOpenLoop)
	}
	if e.NotionOpenLoop.PartialCount == 0 || e.NotionOpenLoop.CadenceHours != readinessEvidenceTTLHours {
		t.Fatalf("notion partial/cadence evidence missing: %+v", e.NotionOpenLoop)
	}
}

func TestOperationalReadinessRejectsNotionCycleWithoutCheck(t *testing.T) {
	m := loadManifestForTest(t)
	m.NotionOpenLoop.Cycles[0].BackfilledCheck = "missing_check"
	if err := verifyNotionOpenLoop("../..", m); err == nil {
		t.Fatal("expected unknown backfilled check to fail")
	}
}
