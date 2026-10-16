package main

import (
	"strings"
	"testing"
	"time"
)

func TestRenewedReviewDeadlineUsesUTCDayAtKSTBoundary(t *testing.T) {
	m := loadDecisionManifestForTest(t)
	decision := m.Decisions[0]
	due, err := time.Parse(reviewDateLayout, decision.ReviewBy)
	if err != nil {
		t.Fatal(err)
	}
	kst := time.FixedZone("KST", 9*60*60)
	for _, check := range []struct {
		name    string
		at      time.Time
		expired bool
	}{
		{"UTC final second", due.Add(24*time.Hour - time.Second), false},
		{"UTC next day", due.Add(24 * time.Hour), true},
		{"KST final second", due.Add(24*time.Hour - time.Second).In(kst), false},
		{"KST next day at 09", due.Add(24 * time.Hour).In(kst), true},
	} {
		t.Run(check.name, func(t *testing.T) {
			t.Setenv("RIIDO_EVIDENCE_NOW", check.at.Format(time.RFC3339))
			err := verifyDecision(decision)
			if check.expired {
				if err == nil || !strings.Contains(err.Error(), "review_by is expired") {
					t.Fatalf("expected UTC-day expiry: %v", err)
				}
			} else if err != nil {
				t.Fatalf("review must remain valid through its UTC date: %v", err)
			}
		})
	}
}
