package main

import (
	"strings"
	"testing"
)

func TestOperationalReadinessDocRendersPartialPromotion(t *testing.T) {
	e := evidence{PartialPromotion: partialPromotion{
		CandidateArtifact: "candidates",
		CandidateCount:    1,
		CandidateIDs:      []string{"operational-readiness:old"},
		StalePartialCount: 1,
		StalePartialIDs:   []string{"old"},
		StaleAfterDays:    2,
	}}
	doc := renderDoc(manifest{Title: "Readiness"}, e)
	if !strings.Contains(doc, "## Partial Promotion") ||
		!strings.Contains(doc, "operational-readiness:old") {
		t.Fatalf("partial promotion missing from doc:\n%s", doc)
	}
}
