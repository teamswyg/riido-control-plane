package main

import (
	"fmt"
	"strings"
	"time"
)

const reviewDateLayout = "2006-01-02"

func verifyDecisionReviewBy(decision decisionRecord) error {
	if !decisionNeedsReviewBy(decision.Disposition) {
		return nil
	}
	if strings.TrimSpace(decision.ReviewBy) == "" {
		return fmt.Errorf("candidate %s needs review_by", decision.CandidateID)
	}
	reviewBy, err := time.Parse(reviewDateLayout, decision.ReviewBy)
	if err != nil {
		return fmt.Errorf("candidate %s review_by must be YYYY-MM-DD", decision.CandidateID)
	}
	today, err := time.Parse(reviewDateLayout, time.Now().UTC().Format(reviewDateLayout))
	if err != nil {
		return fmt.Errorf("parse current review date: %w", err)
	}
	if reviewBy.Before(today) {
		return fmt.Errorf("candidate %s review_by is expired", decision.CandidateID)
	}
	return nil
}

func decisionNeedsReviewBy(disposition string) bool {
	return disposition == "triage_required" || disposition == "deferred"
}
