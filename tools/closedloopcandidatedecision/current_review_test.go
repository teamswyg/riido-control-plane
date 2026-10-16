package main

import (
	"strings"
	"testing"
	"time"
)

func TestPendingDecisionReviewPassesCurrentWallClock(t *testing.T) {
	// Historical producer fixtures pin their own clocks. This policy check must
	// also run against the actual date, including when CI inherited a fixture clock.
	t.Setenv("RIIDO_EVIDENCE_NOW", "")
	m := loadDecisionManifestForTest(t)
	if _, err := verifyAll(repoRootForTest(t), m); err != nil {
		t.Fatalf("current decision policy: %v", err)
	}
	var document struct {
		Review pendingReview `json:"review_evidence"`
	}
	if err := readJSON(repoPath(repoRootForTest(t), defaultManifest), &document); err != nil {
		t.Fatal(err)
	}
	assertPendingReviewDates(t, document.Review)
	assertPendingReviewCoverage(t, m, document.Review)
}

func assertPendingReviewDates(t *testing.T, review pendingReview) {
	t.Helper()
	if review.SchemaVersion != "riido-control-plane-candidate-decision-review.v1" {
		t.Fatal("unexpected pending review identity")
	}
	reviewed, err := time.Parse(reviewDateLayout, review.ReviewedOn)
	if err != nil {
		t.Fatal(err)
	}
	due, err := time.Parse(reviewDateLayout, review.ReviewBy)
	if err != nil {
		t.Fatal(err)
	}
	today, _ := time.Parse(reviewDateLayout, time.Now().UTC().Format(reviewDateLayout))
	if reviewed.After(today) || due.Before(today) || review.IntervalDays != 7 ||
		!due.Equal(reviewed.AddDate(0, 0, review.IntervalDays)) {
		t.Fatal("review must be real, current, and due again within seven days")
	}
}

func TestRenewedReviewStillExpiresDecisionsAndTemplates(t *testing.T) {
	m := loadDecisionManifestForTest(t)
	for _, decision := range []decisionRecord{
		m.Decisions[0], decisionFromTemplate("template:"+m.DecisionTemplates[0].SubjectKind, m.DecisionTemplates[0]),
	} {
		t.Run(decision.CandidateID, func(t *testing.T) {
			due, err := time.Parse(reviewDateLayout, decision.ReviewBy)
			if err != nil {
				t.Fatal(err)
			}
			t.Setenv("RIIDO_EVIDENCE_NOW", due.AddDate(0, 0, 1).Format(time.RFC3339))
			if err := verifyDecision(decision); err == nil || !strings.Contains(err.Error(), "review_by is expired") {
				t.Fatalf("renewed review expiry must fail closed: %v", err)
			}
		})
	}
}
